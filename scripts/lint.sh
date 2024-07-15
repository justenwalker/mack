#!/bin/bash

GOPATH=${GOPATH:-$(go env GOPATH)}
PATH=${GOPATH}/bin:${PATH}

if ! command -v golangci-lint > /dev/null; then
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.1
fi

golangci-lint run --fix
