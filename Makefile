COMPOSE := docker-compose -f docker-compose.yml -f docker-compose.dev.yml
DOCKER_OPTS := -v /var/run/docker.sock:/var/run/docker.sock:ro
GO_FILES = $(shell find ./api -type f -name '*.go' -not -path './api/vendor/*')

.PHONY: default
default: build up
	${COMPOSE} logs --tail 10 -f api

.PHONY: build
build:
	${COMPOSE} build

.PHONY: up
up:
	${COMPOSE} up -d

.PHONY: logs
logs:
	${COMPOSE} logs --tail 50 -f

.PHONY: down
down:
	${COMPOSE} down

.PHONY: pull
pull:
	${COMPOSE} pull

.PHONY: test-short
test-short: test-build
	docker run --rm snip-test go test -v -short

.PHONY: test
test: test-build
	docker run --rm ${DOCKER_OPTS} snip-test go test -v

.PHONY: test-all
test-all: test-build
	docker run --rm ${DOCKER_OPTS} snip-test go test -v -args -languages

.PHONY: test-build
test-build:
	cd api && docker build --target builder -t snip-test .

.PHONY: tool
tool: test-build
	docker run --rm -v $$(pwd):/v -u $$(id -u) snip-test \
		sh -c 'go build -o /tmp/tool ./cmd/tool && cd /v && /tmp/tool'

.PHONY: runner-build
runner-build:
	cd api && docker build -f Dockerfile.runner -t snip-runner-builder .

.PHONY: image-build
image-build:
	cd languages && ./build.sh

.PHONY: image-build-ash
image-build-ash:
	cd languages && ./build.sh ash

.PHONY: go-vet
go-vet:
	@go vet $$(go list ./api/...)

.PHONY: go-fmt
go-fmt:
	@gofmt -l -w $(GO_FILES)

.PHONY: go-simplify
go-simplify:
	@gofmt -s -l -w $(GO_FILES)
