#!/usr/bin/env bash

# Copyright 2019 The cloudsql-postgres-operator Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o nounset
set -o errexit
set -o pipefail

ROOT="${ROOT:-$(git rev-parse --show-toplevel)}"

BINDIR="${ROOT}/bin"

export GO111MODULE="on"

cd "${ROOT}/hack/tools"
go build -o "${BINDIR}/client-gen" k8s.io/code-generator/cmd/client-gen
go build -o "${BINDIR}/deepcopy-gen" k8s.io/code-generator/cmd/deepcopy-gen
go build -o "${BINDIR}/informer-gen" k8s.io/code-generator/cmd/informer-gen
go build -o "${BINDIR}/lister-gen" k8s.io/code-generator/cmd/lister-gen
cd "${ROOT}"
go mod vendor

export GO111MODULE="off"

FAKE_GOPATH="$(mktemp -d)"
trap 'rm -rf ${FAKE_GOPATH}' EXIT

FAKE_REPOPATH="${FAKE_GOPATH}/src/github.com/travelaudience/cloudsql-postgres-operator"
mkdir -p "$(dirname "${FAKE_REPOPATH}")" && ln -s "${ROOT}" "${FAKE_REPOPATH}"

export GOPATH="${FAKE_GOPATH}"
cd "${FAKE_REPOPATH}"

BASE_PACKAGE="github.com/travelaudience/cloudsql-postgres-operator"
CODEGEN_PACKAGES="${BASE_PACKAGE}/pkg/apis/cloudsql/v1alpha1"

"${BINDIR}/deepcopy-gen" \
    --input-dirs ${CODEGEN_PACKAGES} \
    -O zz_generated.deepcopy \
    --bounding-dirs ${BASE_PACKAGE}/pkg/apis \
    --go-header-file "${FAKE_REPOPATH}/hack/header.go.txt"

"${BINDIR}/client-gen" \
    --clientset-name versioned \
    --input-base '' \
    --input ${CODEGEN_PACKAGES} \
    --output-package ${BASE_PACKAGE}/pkg/client/clientset \
    --go-header-file "${FAKE_REPOPATH}/hack/header.go.txt"

"${BINDIR}/lister-gen" \
    --input-dirs ${CODEGEN_PACKAGES} \
    --output-package ${BASE_PACKAGE}/pkg/client/listers \
    --go-header-file "${FAKE_REPOPATH}/hack/header.go.txt"

"${BINDIR}/informer-gen" \
    --input-dirs ${CODEGEN_PACKAGES} \
    --versioned-clientset-package ${BASE_PACKAGE}/pkg/client/clientset/versioned \
    --listers-package ${BASE_PACKAGE}/pkg/client/listers \
    --output-package ${BASE_PACKAGE}/pkg/client/informers \
    --go-header-file "${FAKE_REPOPATH}/hack/header.go.txt"

export GO111MODULE="on"

cd "${ROOT}"
