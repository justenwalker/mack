#!/bin/bash

GOPATH=${GOPATH:-$(go env GOPATH)}
PATH=${GOPATH}/bin:${PATH}

if ! command -v gci > /dev/null; then
  go install github.com/daixiang0/gci@v0.13.4
fi

if ! command -v gofumpt > /dev/null; then
  go install mvdan.cc/gofumpt@v0.6.0
fi

gci write -s standard -s default -s "prefix(github.com/justenwalker/mack)" ./
gofumpt -l -w .
