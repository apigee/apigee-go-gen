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


sortSemver() {
  local lines=""
  while read version; do
    if [[ -z "${lines}" ]] ; then
      lines=$(printf '%s' "${version}")
    else
      lines=$(printf '%s\n%s' "${lines}" "${version}")
    fi
  done
  echo "$lines" | sed -r 's:^v::' | sed -r 's:-:~:' | sort -r -V | sed -r 's:^:v:' | sed -r 's:~:-:'
}

pickLatestRelease() {
  local first=""
  while read version; do
    if [[ -z "${first}" ]] ; then
      first="${version}"
    fi
    if [[ "${version}" != *"-"* ]] ; then
      echo "${version}"
      return
    fi
  done
  echo "${first}"
}

getReleasedTags() {
  git tag --list  | grep "^v"
}

getLatestRelease() {
  echo "$(getReleasedTags | sortSemver | pickLatestRelease)"
}

REGISTRY="${REGISTRY:-ghcr.io}"
GIT_REPO="${GIT_REPO:-apigee/apigee-go-gen}"
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_TAG=$(git describe --tags --abbrev=0)

if [[ "$(getLatestRelease)" == "${GIT_TAG}" ]] ; then
  BUILD_TAG="latest"
else
  BUILD_TAG="${GIT_TAG}"
fi

echo "BUILD_TAG=${BUILD_TAG}"
echo "GIT_TAG=${GIT_TAG}"

docker buildx create --name builder --use

OCI="index:org.opencontainers.image"
docker buildx build  \
       --platform=linux/amd64,linux/arm64  \
       --tag "${REGISTRY}/${GIT_REPO}:${GIT_TAG}" \
       --tag "${REGISTRY}/${GIT_REPO}:${BUILD_TAG}" \
       --provenance false \
       --output type=registry \
       --annotation "${OCI}.url=https://github.com/${GIT_REPO}" \
       --annotation "${OCI}.documentation=https://github.com/${GIT_REPO}" \
       --annotation "${OCI}.source=https://github.com/${GIT_REPO}" \
       --annotation "${OCI}.version=${GIT_TAG}" \
       --annotation "${OCI}.revision=${GIT_COMMIT}" \
       --annotation "${OCI}.vendor=Google LLC" \
       --annotation "${OCI}.licenses=Apache-2.0" \
       --annotation "${OCI}.description=This is a tool for generating Apigee bundles and shared flows" \
       --push \
       .

docker buildx rm builder
