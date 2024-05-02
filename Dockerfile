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

FROM --platform=$BUILDPLATFORM golang:alpine as builder
RUN apk update && apk add --no-cache git

ADD ./ /src
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

RUN sh build.sh

FROM --platform=$TARGETPLATFORM alpine:3
COPY LICENSE /
COPY --from=builder /src/bin/* /usr/local/bin/
ENTRYPOINT [ "apigee-go-gen" ]

