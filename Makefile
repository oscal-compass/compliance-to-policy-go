# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GIT_TAG := $(shell git describe --tags --abbrev=0)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
VERSION_FROM_GIT_TAG := $(shell echo "$(GIT_TAG)" | sed 's/^go\///')
DIRTY := $(shell [ -n "$(git status -s)" ] && echo '-snapshot')
REPO_NAME := $(shell git remote get-url origin | sed -r 's/.*:(.*)\.git/\1/')
VERSIONED_SUFFIX := $(if $(DIRTY),$(VERSION_FROM_GIT_TAG)_$(GOOS)_$(GOARCH),$(VERSION_FROM_GIT_TAG)_SNAPSHOT_$(GOOS)_$(GOARCH))

GO_LD_EXTRAFLAGS := -X github.com/oscal-compass/compliance-to-policy-go/v2/cmd/c2pcli/cli/subcommands.version="$(GIT_TAG)" \
                    -X github.com/oscal-compass/compliance-to-policy-go/v2/cmd/c2pcli/cli/subcommands.commit="$(GIT_COMMIT)" \
                    -X github.com/oscal-compass/compliance-to-policy-go/v2/cmd/c2pcli/cli/subcommands.date="$(BUILD_DATE)"

repo_name:
	echo $(REPO_NAME)

.PHONY: all
all: build

.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/c2pcli_$(VERSIONED_SUFFIX) ./cmd/c2pcli

.PHONY: build-plugins
build-plugins:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/kyverno-plugin ./cmd/kyverno-plugin
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/ocm-plugin ./cmd/ocm-plugin

.PHONY: test
test:
	go test ./... -coverprofile cover.out

artifact: build
	mkdir -p ./dist/artifacts
	tar zcvf ./dist/artifacts/c2pcli_$(VERSIONED_SUFFIX).tar.gz -C ./bin c2pcli_$(VERSIONED_SUFFIX)
	shasum -a 256 ./dist/artifacts/c2pcli_$(VERSIONED_SUFFIX).tar.gz > ./dist/artifacts/c2pcli_$(VERSIONED_SUFFIX).sha256

# echo $PAT | gh auth login --with-token -h github.com
release: GITHUB_HOST ?= github.com
release: artifact
	@(gh release --repo $(GITHUB_HOST)/$(REPO_NAME) view $(GIT_TAG) ;\
	if [[ "$$?" != "0" ]];then \
		echo create release $(GIT_TAG) ;\
		gh release --repo $(GITHUB_HOST)/$(REPO_NAME) create $(GIT_TAG) --generate-notes ;\
	fi)
	gh release --repo $(GITHUB_HOST)/$(REPO_NAME) upload $(GIT_TAG) ./dist/artifacts/c2pcli_$(VERSIONED_SUFFIX).*

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
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

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...


##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.7

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

generate-protobuf:
	protoc api/proto/*.proto --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_out=. --go_opt=paths=source_relative --proto_path=.
.PHONY: generate-protobuf

