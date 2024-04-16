#!/bin/bash

mkdir -p bin
echo "Building tool ..."
go build -o bin/proxy-tool ./cmd/proxy-tool

