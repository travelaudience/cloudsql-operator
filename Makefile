# SHELL specifies which shell to use.
SHELL := /bin/bash

# ROOT holds the absolute path to the root of the cloudsql-operator repository.
ROOT := $(shell git rev-parse --show-toplevel)

# VERSION holds the version of cloudsql-operator being built.
VERSION ?= $(shell git describe --always --dirty=-dev)

# build builds the cloudsql-operator binary for the specified architecture (defaults to "amd64") and operating system (defaults to "linux").
.PHONY: build
build: gen
build: GOARCH ?= amd64
build: GOOS ?= linux
build:
	@GOARCH=$(GOARCH) GOOS=$(GOOS) go build \
		-ldflags="-s -w -X github.com/travelaudience/cloudsql-operator/pkg/version.Version=$(VERSION)" \
		-o $(ROOT)/build/cloudsql-operator \
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
		github.com/travelaudience/cloudsql-operator/pkg/client \
		github.com/travelaudience/cloudsql-operator/pkg/apis \
		cloudsql:v1alpha1 \
		--go-header-file $(ROOT)/hack/header.go.txt

# skaffold deploys cloudsql-operator to the Kubernetes cluster targeted by the current context.
.PHONY: skaffold
skaffold: MODE ?= dev
skaffold: PROFILE ?= minikube
skaffold:
	@if [[ ! "$(MODE)" == "delete" ]]; then \
		GOOS=linux GOARCH=amd64 $(MAKE) -C $(ROOT) build; \
	fi
	@skaffold $(MODE) --filename $(ROOT)/hack/skaffold/skaffold.yaml --profile $(PROFILE)
