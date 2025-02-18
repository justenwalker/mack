#!/bin/bash

go tool -modfile tools.mod github.com/daixiang0/gci write -s standard -s default -s "prefix(github.com/justenwalker/mack)" ./
go tool -modfile tools.mod mvdan.cc/gofumpt -l -w .
