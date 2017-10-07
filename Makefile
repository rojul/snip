COMPOSE := docker-compose -f docker-compose.yml -f docker-compose.dev.yml

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

runner-build:
	cd api && docker build -f Dockerfile.runner -t snip-runner-builder .

image-build:
	cd languages && ./build.sh

image-build-ash:
	cd languages && ./build.sh ash

.PHONY: default build up logs down pull runner-build image-build image-build-ash
