create table main.settings
(
    name  text,
    value text
) strict;
create table main.projects
(
    id        integer primary key autoincrement,
    name      text,
    pri       integer default 2 not null, -- user-defined priority of project (0-3)
    open      integer default 0 not null, -- num of open tasks (deferred excluded)
    closed    integer default 0 not null, -- num of resolved tasks
    closed_mo integer default 0 not null, -- num of tasks resolved in last 30d
    time      integer default 0 not null, -- total work time
    time_mo   integer default 0 not null, -- sum of work time in last 30d
    rank      real    default 0 not null, -- abstract rank of project (calc in trigger on task close)
    urgency   real generated always as (0.1 + rank * (4 - pri) / 4) stored
) strict;

create table main.tasks
(
    uid        integer primary key autoincrement,
    id         integer default 0,
    pri        integer default 2, -- user-defined priority of task (0-3)
    desc       text,
    project_id int     default 0                                    not null,
    created    int     default (cast(strftime('%s', 'now') as int)) not null,
    updated    int     default (cast(strftime('%s', 'now') as int)) not null,
    started    int     default 0                                    not null,
    closed     int     default 0                                    not null,
    deferred   int     default 0                                    not null,
    time       int     default 0                                    not null,
    due        int     default 0                                    not null
) strict;
create index tasks_id_index on tasks (id);
create index tasks_closed_index on tasks (closed);

create trigger set_id_after_insert
    after insert
    on tasks
begin
    update tasks set id = (select max(id) + 1 from tasks) where uid = new.uid;
end;

create trigger up_proj_after_insert
    after insert
    on tasks
    when new.project_id != 0
begin
    update projects
    set open = (select count(uid) from tasks where closed = 0 and deferred = 0 and project_id = new.project_id)
    where id = new.project_id;
end;

create trigger set_updated_after_update
    after update
    on tasks
    when new.updated IS old.updated
begin
    update tasks set updated = unixepoch('now') where uid = new.uid;
end;

create trigger up_projects_after_task_update
    after update
    on tasks
    when old.project_id != new.project_id
begin
    update projects
    set open = (select count(uid) from tasks where closed = 0 and project_id = new.project_id)
    where id = new.project_id;
    update projects
    set open = (select count(uid) from tasks where closed = 0 and project_id = old.project_id)
    where id = old.project_id;
end;

create trigger task_after_close
    after update
    on tasks
    when old.closed is not new.closed and new.closed != 0
begin
    update tasks set id = null, closed = unixepoch('now') where uid = new.uid;
end;

create trigger up_project_after_task_close
    after update
    on tasks
    when new.project_id != 0 and old.closed is not new.closed and new.closed != 0
begin
    update projects
    set open      = sub.open,
        closed    = sub.closed,
        closed_mo = sub.closed_month,
        time_mo   = sub.time_month,
        time      = sub.time
    from (select sum(case when closed = 0 AND deferred = 0 then 1 else 0 end)                    as open,
                 sum(case when closed != 0 then 1 else 0 end)                                    as closed,
                 sum(case when closed > unixepoch(date('now', '-1 month')) then 1 else 0 end)    as closed_month,
                 sum(case when closed > unixepoch(date('now', '-1 month')) then time else 0 end) as time_month,
                 sum(time)                                                                       as time
          from tasks
          where project_id = new.project_id) as sub
    where id = new.project_id;

    update projects
    set rank = round((timeRank + openRank + closedRank) / 3, 2)
    from (select id,
                 percent_rank() over (order by time_mo desc)   as timeRank,
                 percent_rank() over (order by open)           as openRank,
                 percent_rank() over (order by closed_mo desc) as closedRank
          from projects) as s
    where s.id = projects.id;
end;

create trigger calc_time_spent
    after update
    on tasks
    when old.started != 0 and new.started = 0
begin
    insert into times(task_uid, start, end) values (new.uid, old.started, unixepoch('now'));
    update tasks set time = (select sum(duration) from times where task_uid = new.uid) where uid = new.uid;
end;

create table main.tags_tasks
(
    tag_id   integer,
    task_uid integer references tasks on delete cascade,
    constraint tags_tasks_pk unique (tag_id, task_uid)
) strict;
create table main.tags
(
    id   integer not null primary key autoincrement,
    name text
) strict;
create table main.times
(
    task_uid integer references tasks on delete cascade,
    start    integer,
    end      integer,
    duration integer generated always as (end - start) stored
) strict;

create view tasks_view as
select tasks.uid                                    as uid,
       tasks.id                                     as id,
       tasks.pri                                    as pri,
       tasks.desc                                   as desc,
       tasks.project_id                             as project_id,
       ifnull(projects.name, '')                    as project,
       ifnull(projects.urgency, 0)                  as project_urgency,
       tasks.created                                as created,
       tasks.started                                as started,
       tasks.closed                                 as closed,
       tasks.time                                   as time,
       tasks.due                                    as due,
       tasks.deferred                               as deferred,
       ifnull(group_concat(distinct tags.name), '') as tags,
       count(distinct tags.name)                    as tags_count
from tasks
         left join projects on tasks.project_id = projects.id
         left outer join tags_tasks on tasks.uid = tags_tasks.task_uid
         left outer join tags on tags_tasks.tag_id = tags.id
group by tasks.id;
