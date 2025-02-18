#!/bin/bash

go mod tidy
go mod -modfile tools.mod tidy
cd ./example/ || exit
go mod tidy