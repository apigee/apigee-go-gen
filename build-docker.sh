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
    first="${version}"
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
BUILD_TIMESTAMP=$(date "+%s")
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_TAG=$(git describe --tags --abbrev=0)
LATEST_TAG="$(getLatestRelease)"

if [[ "${LATEST_TAG}" == "${GIT_TAG}" ]] ; then
  BUILD_TAG="latest"
else
  BUILD_TAG="${GIT_TAG}"
fi


echo "LATEST_TAG=${LATEST_TAG}"
echo "BUILD_TAG=${BUILD_TAG}"
echo "GIT_TAG=${GIT_TAG}"
echo "GIT_COMMIT=${GIT_COMMIT}"

docker build -t "${REGISTRY}/${GIT_REPO}:${BUILD_TAG}" \
       -t "${REGISTRY}/${GIT_REPO}:${GIT_TAG}" \
       -t "${REGISTRY}/${GIT_REPO}:${GIT_COMMIT}" \
       --build-arg "GIT_REPO=${GIT_REPO}" \
       --build-arg="GIT_TAG=${GIT_TAG}" \
       --build-arg="GIT_COMMIT=${GIT_COMMIT}" \
       --build-arg="BUILD_TIMESTAMP=${BUILD_TIMESTAMP}" \
       .

if [ "${1}" == "push" ] ; then
  docker push "${REGISTRY}/${GIT_REPO}:${BUILD_TAG}"
  docker push "${REGISTRY}/${GIT_REPO}:${GIT_TAG}"
  docker push "${REGISTRY}/${GIT_REPO}:${GIT_COMMIT}"
fi
