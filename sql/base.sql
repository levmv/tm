create table main.settings
(
    name  text,
    value text
) strict;
create table main.projects
(
    id         integer primary key autoincrement,
    name       text,
    priority   integer default 2 not null,
    open       integer default 0 not null,
    closed     integer default 0 not null,
    time_total integer default 0 not null
) strict;

create table main.tasks
(
    uid        integer primary key autoincrement,
    id         integer default 0,
    state      integer,
    priority   integer default 2,
    summary    text,
    project_id int     default 0                                    not null,
    created    int     default (cast(strftime('%s', 'now') as int)) not null,
    updated    int     default (cast(strftime('%s', 'now') as int)) not null,
    started    int     default 0                                    not null,
    closed     int     default 0                                    not null,
    deferred   int     default 0                                    not null,
    time_spent int     default 0                                    not null,
    due        int     default 0                                    not null
) strict;
create index tasks_id_index on tasks (id);
create index tasks_state_index on tasks (state);

create trigger set_id_after_insert
    after insert
    on tasks
begin
    update tasks set id = (select max(id) + 1 from tasks) where uid = new.uid;
end;
create trigger set_updated_after_update
    after update
    on tasks
    when new.updated IS old.updated
begin
    update tasks set updated = unixepoch('now') where uid = new.uid;
end;
-- TODO: do we really need to update counters or maybe view will be enough?
create trigger up_projects_after_task_update
    after update
    on tasks
    when old.project_id != new.project_id
begin
    update projects
    set
        open = (select count(uid) from tasks where state = 0 and project_id = new.project_id)
    where id = new.project_id;
    update projects
    set
        open = (select count(uid) from tasks where state = 0 and project_id = old.project_id)
    where id = old.project_id;
end;
create trigger task_after_close
    after update
    on tasks
    when old.state is not new.state and new.state == 1
begin
    update tasks set id = null, closed = unixepoch('now') where uid = new.uid;
end;
create trigger up_project_after_task_close
    after update
    on tasks
    when new.project_id != 0 and old.state is not new.state and new.state == 1
begin
    update projects
    set open = (select count(*) from tasks where state = 0 and project_id = new.project_id)
    where id = new.project_id;

    update projects
    set closed = (select count(*) from tasks where state = 1 and project_id = new.project_id)
    where id = new.project_id;

    update projects
    set time_total = (select sum(time_spent) from tasks where state = 1 and project_id = new.project_id)
    where id = new.project_id;
end;
create trigger calc_time_spent
    after update
    on tasks
    when old.started != 0 and new.started = 0
begin
    insert into times(task_id, start, end) values (new.uid, old.started, unixepoch('now'));
    update tasks set time_spent = (select sum(duration) from times where task_id = new.uid) where uid = new.uid;
end;

create table main.tags_tasks
(
    tag_id  integer,
    task_id integer references tasks on delete cascade,
    constraint tags_tasks_pk unique (tag_id, task_id)
) strict;
create table main.tags
(
    id   integer not null primary key autoincrement,
    name text
) strict;
create table main.times
(
    task_id  integer references tasks on delete cascade,
    start    integer,
    end      integer,
    duration integer generated always as (end - start) stored
) strict;

create view tasks_view as
select tasks.uid                                    as uid,
       tasks.id                                     as id,
       tasks.state                                  as state,
       tasks.priority                               as priority,
       tasks.summary                                as summary,
       tasks.project_id                             as project_id,
       ifnull(projects.name, '')                    as project,
       ifnull(projects.priority, 0)                 as project_priority,
       tasks.created                                as created,
       tasks.started                                as started,
       tasks.time_spent                             as time_spent,
       tasks.due                                    as due,
       tasks.deferred                               as deferred,
       ifnull(group_concat(distinct tags.name), '') as tags,
       count(distinct tags.name)                    as tags_count
from tasks
         left join projects on tasks.project_id = projects.id
         left outer join tags_tasks on tasks.uid = tags_tasks.task_id
         left outer join tags on tags_tasks.tag_id = tags.id
group by tasks.id;
