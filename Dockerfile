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

FROM golang:1.21 as builder

ARG GIT_REPO
ARG GIT_TAG
ARG GIT_COMMIT
ARG BUILD_TIMESTAMP

ADD ./ /src
WORKDIR /src
RUN ./build.sh

FROM golang:1.21

COPY LICENSE /

COPY --from=builder /src/bin/* /usr/local/bin/

ARG GIT_REPO
ARG GIT_TAG
ARG GIT_COMMIT
ARG BUILD_TIMESTAMP

LABEL org.opencontainers.image.url="https://github.com/${GIT_REPO}" \
      org.opencontainers.image.documentation="https://github.com/${GIT_REPO}" \
      org.opencontainers.image.source="https://github.com/${GIT_REPO}" \
      org.opencontainers.image.version="${GIT_TAG}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.vendor='Google LLC' \
      org.opencontainers.image.licenses='Apache-2.0' \
      org.opencontainers.image.description='This is a tool for generating Apigee bundles and shared flows'

ENTRYPOINT [ "apigee-go-gen" ]

