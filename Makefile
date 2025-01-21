
# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/dfradehubs/akapurgo:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OS=$(shell uname | tr '[:upper:]' '[:lower:]')

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.54.2
golangci-lint:
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: check-go-target
check-go-target: ## Check presente of GOOS and GOARCH vars.
	@if [ -z "$(GOOS)" ]; then \
		echo "GOOS is not defined. Define GOOS y try again."; \
		exit 1; \
	fi
	@if [ -z "$(GOARCH)" ]; then \
		echo "GOARCH is not defined. Define GOARCH y try again."; \
		exit 1; \
	fi

.PHONY: build
build: fmt vet check-go-target ## Build CLI binary.
	go build -o bin/akapurgo-$(GOOS)-$(GOARCH) cmd/main.go

.PHONY: run
run: fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-builder
	$(CONTAINER_TOOL) buildx use project-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm project-builder
	rm Dockerfile.cross


PACKAGE_NAME ?= package.tar.gz
.PHONY: package
package: check-go-target ## Package binary.
	@printf "\nCreating package at dist/$(PACKAGE_NAME) \n"
	@mkdir -p dist

	@if [ "$(OS)" = "linux" ]; then \
		tar --transform="s/akapurgo-$(GOOS)-$(GOARCH)/akapurgo/" -cvzf dist/$(PACKAGE_NAME) -C bin akapurgo-$(GOOS)-$(GOARCH) -C ../ LICENSE README.md; \
	elif [ "$(OS)" = "darwin" ]; then \
		tar -cvzf dist/$(PACKAGE_NAME) -s '/akapurgo-$(GOOS)-$(GOARCH)/akapurgo/' -C bin akapurgo-$(GOOS)-$(GOARCH) -C ../ LICENSE README.md; \
	else \
		echo "Unsupported OS: $(GOOS)"; \
		exit 1; \
	fi

.PHONY: package-signature
package-signature: ## Create a signature for the package.
	@printf "\nCreating package signature at dist/$(PACKAGE_NAME).md5 \n"
	md5sum dist/$(PACKAGE_NAME) | awk '{ print $$1 }' > dist/$(PACKAGE_NAME).md5

.PHONY: run-example
run-example:
	@echo "Creating example environment..."
	docker-compose -f config/samples/docker-compose.yml up --build