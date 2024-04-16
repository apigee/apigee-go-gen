#!/bin/bash

export VERSION="0.1.5"
export BUILD_TIMESTAMP="$(date "+%s")"
export BUILD_GIT_HASH="$(git show -s --format=%h)"
export LD_FLAGS="-X main.BuildVersion=${VERSION} -X main.BuildTimestamp=${BUILD_TIMESTAMP} -X main.BuildGitHash=${BUILD_GIT_HASH}"


mkdir -p bin
echo "Building templating tools ..."
go build -ldflags "${LD_FLAGS}" -o bin/render-template  ./cmd/render-template
go build -ldflags "${LD_FLAGS}" -o bin/render-bundle  ./cmd/render-bundle



echo "Building transformation tools ..."
go build -ldflags "${LD_FLAGS}" -o bin/yaml2bundle  ./cmd/yaml2bundle
go build -ldflags "${LD_FLAGS}" -o bin/bundle2yaml  ./cmd/bundle2yaml
go build -ldflags "${LD_FLAGS}" -o bin/xml2yaml  ./cmd/xml2yaml
go build -ldflags "${LD_FLAGS}" -o bin/yaml2xml  ./cmd/yaml2xml


