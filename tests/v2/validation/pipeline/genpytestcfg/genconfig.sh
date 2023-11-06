#!/bin/bash
set -e
cd $(dirname $0)/../../../../../

echo "building release upgrade bin"
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o tests/v2/validation/pipeline/bin/genconfig ./tests/v2/validation/pipeline/genvalidationtestcfg

echo "running release upgrade"
tests/v2/validation/pipeline/bin/genconfig
