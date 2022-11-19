clean:
	docker rmi cody:latest

install_dependencies:
	go install github.com/mitranim/gow@latest

test_watch:
	gow -s test -v ./...

test:
	go test -v ./...