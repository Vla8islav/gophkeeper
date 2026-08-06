.PHONY: generate test

generate:
	go generate ./internal/mocks

test:
	go test ./...
