COMPOSE := docker-compose -f docker-compose.yml -f docker-compose.dev.yml
DOCKER_OPTS := -v /var/run/docker.sock:/var/run/docker.sock:ro

default: build up
	${COMPOSE} logs --tail 10 -f api

build:
	${COMPOSE} build

up:
	${COMPOSE} up -d

logs:
	${COMPOSE} logs --tail 50 -f

down:
	${COMPOSE} down

pull:
	${COMPOSE} pull

test: test-build
	docker run --rm ${DOCKER_OPTS} snip-test go test -v -args -languages

test-short: test-build
	docker run --rm snip-test go test -v -short

test-build:
	cd api && docker build --target builder -t snip-test .

runner-build:
	cd api && docker build -f Dockerfile.runner -t snip-runner-builder .

image-build:
	cd languages && ./build.sh

image-build-ash:
	cd languages && ./build.sh ash

.PHONY: default build up logs down pull test test-short test-build runner-build image-build image-build-ash
