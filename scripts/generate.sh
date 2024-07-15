#!/bin/bash

go generate ./...
docker compose -f ./docker/docker-compose.yaml run --no-TTY --rm --build protogen