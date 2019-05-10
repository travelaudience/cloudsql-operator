# SHELL specifies which shell to use.
SHELL := /bin/bash

# ROOT holds the absolute path to the root of the cloudsql-postgres-operator repository.
ROOT := $(shell git rev-parse --show-toplevel)

# VERSION holds the version of cloudsql-postgres-operator being built.
VERSION ?= $(shell git describe --always --dirty=-dev)

# build builds the cloudsql-postgres-operator binary for the specified architecture (defaults to "amd64") and operating system (defaults to "linux").
.PHONY: build
build: gen
build: GOARCH ?= amd64
build: GOOS ?= linux
build:
	@GOARCH=$(GOARCH) GOOS=$(GOOS) go build \
		-ldflags="-s -w -X github.com/travelaudience/cloudsql-postgres-operator/pkg/version.Version=$(VERSION)" \
		-o $(ROOT)/build/cloudsql-postgres-operator \
		-v \
		$(ROOT)/cmd/operator/main.go

# dep installs the project's build dependencies.
.PHONY: dep
dep: KUBERNETES_VERSION=1.13.4
dep: KUBERNETES_CODE_GENERATOR_PKG=k8s.io/code-generator
dep:
	@dep ensure -v
	@go get -d $(KUBERNETES_CODE_GENERATOR_PKG)/... && \
		cd $(GOPATH)/src/$(KUBERNETES_CODE_GENERATOR_PKG) && \
		git fetch origin && \
		git checkout -fq kubernetes-$(KUBERNETES_VERSION)

# gen executes the code generation step.
.PHONY: gen
gen: dep
	@$(GOPATH)/src/k8s.io/code-generator/generate-groups.sh "deepcopy,client,informer,lister" \
		github.com/travelaudience/cloudsql-postgres-operator/pkg/client \
		github.com/travelaudience/cloudsql-postgres-operator/pkg/apis \
		cloudsql:v1alpha1 \
		--go-header-file $(ROOT)/hack/header.go.txt

# skaffold deploys cloudsql-postgres-operator to the Kubernetes cluster targeted by the current context.
.PHONY: skaffold
skaffold: ADMIN_KEY_JSON_FILE ?= $(ROOT)/admin-key.json
skaffold: CLIENT_KEY_JSON_FILE ?= $(ROOT)/client-key.json
skaffold: MODE ?= dev
skaffold: PROFILE ?= minikube
skaffold: PROJECT_ID ?= cloudsql-postgres-operator
skaffold:
	@ADMIN_KEY_JSON_FILE=$(ADMIN_KEY_JSON_FILE) \
	CLIENT_KEY_JSON_FILE=$(CLIENT_KEY_JSON_FILE) \
	MODE=$(MODE) \
	PROFILE=$(PROFILE) \
	PROJECT_ID=$(PROJECT_ID) \
	$(ROOT)/hack/skaffold/skaffold.sh

# test.e2e runs the end-to-end test suite.
.PHONY: test.e2e
test.e2e: FOCUS ?= .*
test.e2e: KUBECONFIG ?= $(HOME)/.kube/config
test.e2e: LOG_LEVEL ?= info
test.e2e: NAMESPACE ?= cloudsql-postgres-operator
test.e2e: NETWORK ?= default
test.e2e: PATH_TO_ADMIN_KEY ?=
test.e2e: PROJECT_ID ?= cloudsql-postgres-operator
test.e2e: REGION ?= europe-west1
test.e2e: TEST_PRIVATE_IP_ACCESS ?= false
test.e2e: TIMEOUT ?= 1800s
test.e2e:
	@go test -tags e2e $(ROOT)/test/e2e \
		-ginkgo.focus="$(FOCUS)" \
		-ginkgo.v \
		-test.timeout="$(TIMEOUT)" \
		-test.v \
		-kubeconfig="$(KUBECONFIG)" \
		-log-level="$(LOG_LEVEL)" \
		-network="$(NETWORK)" \
		-path-to-admin-key="$(shell realpath $(PATH_TO_ADMIN_KEY))" \
		-project-id="$(PROJECT_ID)" \
		-region="$(REGION)" \
		-test-private-ip-access="$(TEST_PRIVATE_IP_ACCESS)"
