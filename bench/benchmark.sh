#!/bin/bash
set -euo pipefail

go test -bench=. -run='^$' -v -count=10 ./... | tee benchmarks.out
benchstat -col /impl benchmarks.out | tee benchstat.out
