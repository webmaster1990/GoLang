all: main

.PHONY: main clean test bench uninstall

main:
	goimports -w .
	go build

install:
	go install
	sudo cp lucid_backend.service /etc/systemd/system/
	sudo systemctl daemon-reload

uninstall:
	go clean -i

clean:
	go clean

test:
	go test ./...

bench:
	go test -bench=. -benchmem
