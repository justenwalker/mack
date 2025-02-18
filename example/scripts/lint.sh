#!/bin/bash

go tool -modfile tools.mod github.com/golangci/golangci-lint/cmd/golangci-lint run --fix .
