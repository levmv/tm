all: test build

test:
	go test

install:
	sudo cp bin/tm-linux-amd64 /usr/local/bin/tm

build:
	go build -ldflags="-w -s" -o bin/tm-linux-amd64
	sha256sum bin/tm-linux-amd64 > bin/tm.linux-amd64.sha256sum