#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

export GO111MODULE=on

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
PACKAGE_BASE="github.com/QubitProducts/prom-config-controller"

${SCRIPT_ROOT}/hack/generate-groups.sh "deepcopy,client,informer,lister" \
  "${PACKAGE_BASE}/pkg/client" \
  "${PACKAGE_BASE}/pkg/apis" \
  config:v1beta1 \
  --output-base "$(dirname "${BASH_SOURCE[0]}")/codegen" \
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt


cp -r \
  "$(dirname "${BASH_SOURCE[0]}")/codegen/${PACKAGE_BASE}/pkg" \
  $SCRIPT_ROOT

rm -rf "$(dirname "${BASH_SOURCE[0]}")/codegen"
