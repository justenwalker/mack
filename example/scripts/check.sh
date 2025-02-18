#!/bin/bash

go vet ./..
go tool -modfile tools.mod golang.org/x/vuln/cmd/govulncheck -show verbose ./...