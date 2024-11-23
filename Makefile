.PHONY: default
default: build

.PHONY: test
test:
	go test -race ./...
	golangci-lint run ./...

.PHONY: build
build: test
	go install .
