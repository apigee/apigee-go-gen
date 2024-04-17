#!/bin/bash
#
#  Copyright 2024 Google LLC
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#


BUILD_TIMESTAMP=$(date "+%s")
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_TAG=$(git describe --tags --abbrev=0)
LD_FLAGS="-X main.GitTag=${GIT_TAG} -X main.BuildTimestamp=${BUILD_TIMESTAMP} -X main.GitCommit=${GIT_COMMIT}"


mkdir -p bin
echo "Building tool ..."
go build -buildvcs=true -ldflags "${LD_FLAGS}" -o bin/apigee-go-gen ./cmd/apigee-go-gen

