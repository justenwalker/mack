#!/bin/bash

GOPATH=${GOPATH:-$(go env GOPATH)}
PATH=${GOPATH}/bin:${PATH}

go install golang.org/x/vuln/cmd/govulncheck@latest

go vet ./..
govulncheck -show verbose ./...