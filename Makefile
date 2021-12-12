.PHONY: *

MAIN := .
TEST := ./internal

test:
	go test -v $(TEST)

test-docker: build-docker
	kustomize build example --enable-alpha-plugins

test-watch:
	watch -n1 go test -v $(TEST)

test-cover:
	go test -coverprofile=coverage.out $(TEST)
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

build:
	goreleaser build --rm-dist --snapshot

build-docker:
	goreleaser build --rm-dist --snapshot
	cp Dockerfile dist/kustomize-variable-injector_linux_amd64
	docker build -t ghcr.io/choffmeister/kustomize-variable-injector:latest dist/kustomize-variable-injector_linux_amd64

release:
	goreleaser release --rm-dist
