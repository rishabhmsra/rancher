#!/bin/bash
set -e
cd $(dirname $0)/../../../../../

echo "building genamiconfigs bin"
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o tests/v2/validation/pipeline/bin/genamiconfigs ./tests/v2/validation/pipeline/genamiconfigs

echo "running genamiconfigs"
tests/v2/validation/pipeline/bin/genamiconfigs