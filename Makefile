.PHONY: build test doc lint bench
build:
	go build ./...

test:
	go test -race ./...

doc:
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc
	gomarkdoc --output '{{.Dir}}/README.md' ./...

lint:
	golangci-lint run

bench:
	go test -run=^$ -bench=. -benchmem ./...
