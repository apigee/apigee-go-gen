#!/bin/bash

BUILD_TIMESTAMP=$(date "+%s")
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_TAG=$(git describe --tags --abbrev=0)
LD_FLAGS="-X main.GitTag=${GIT_TAG} -X main.BuildTimestamp=${BUILD_TIMESTAMP} -X main.GitCommit=${GIT_COMMIT}"


mkdir -p bin
echo "Building tool ..."
go build -buildvcs=true -ldflags "${LD_FLAGS}" -o bin/proxy-tool ./cmd/proxy-tool

