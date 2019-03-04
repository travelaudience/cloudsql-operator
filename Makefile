# ROOT holds the absolute path to the root of the cloudsql-operator repository.
ROOT := $(shell git rev-parse --show-toplevel)

# VERSION holds the version of cloudsql-operator being built.
VERSION ?= $(shell git describe --always --dirty=-dev)

# build builds the cloudsql-operator binary for the specified architecture (defaults to "amd64") and operating system (defaults to "linux").
.PHONY: build
build: dep
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
dep:
	@dep ensure -v
