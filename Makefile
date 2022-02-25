.PHONY: *

test:
	go test -v ./...

test-docker: build
	kustomize build example --enable-alpha-plugins

test-watch:
	watch -n1 go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

build:
	goreleaser release --rm-dist --snapshot --skip-publish

release:
	goreleaser release --rm-dist
