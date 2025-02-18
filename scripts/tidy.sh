#!/bin/bash

go mod tidy
go mod tidy -modfile tools.mod
cd ./example/ || exit
go mod tidy