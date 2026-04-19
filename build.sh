#!/bin/bash
set -e

RUN_NAME=hertz_service
mkdir -p output/bin
mkdir -p output/docs
cp script/* output
cp -r docs/. output/docs
chmod +x output/bootstrap.sh
CGO_ENABLED=0 go build -o output/bin/${RUN_NAME}
