# Copyright 2019 The Kubernetes Authors.
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

.EXPORT_ALL_VARIABLES:

# settings
REGISTRY?=registry.cn-hangzhou.aliyuncs.com/acs/cloud-controller-manager-amd64
MULTI_ARCH_REGISTRY?=registry.cn-hangzhou.aliyuncs.com/acs/cloud-controller-manager
TAG?=$(shell git describe --tags)

REPO_ROOT:=${CURDIR}
OUT_DIR=$(REPO_ROOT)/bin
KIND_BINARY_NAME?=cloud-controller-manager

# go
TARGETOS?=linux
TARGETARCH?=amd64
TARGETPLATFORM?=linux/amd64
GOPROXY?=https://goproxy.cn,direct

# ldflags
VERSION_PKG=k8s.io/cloud-provider-alibaba-cloud/version
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date +%Y-%m-%dT%H:%M:%S%z)
ldflags="-s -w -X $(VERSION_PKG).Version=$(TAG) -X $(VERSION_PKG).GitCommit=${GIT_COMMIT} -X ${VERSION_PKG}.BuildDate=${BUILD_DATE}"

#tools
GOLANGCI_LINT = go run ${REPO_ROOT}/vendor/github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: cloud-controller-manager
cloud-controller-manager: gofmt unit-test
	@echo + Building cloud-controller-manager binary
	GOARCH=${TARGETARCH} \
	GOOS=${TARGETOS} \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	GOPROXY=${GOPROXY} \
	go build -mod vendor -v -o $(OUT_DIR)/$(KIND_BINARY_NAME) \
		    -ldflags $(ldflags) cmd/manager/main.go
	@echo + Built cloud-controller-manager binary to $(OUT_DIR)/$(KIND_BINARY_NAME)

.PHONY: image
image:
ifeq ($(TARGETPLATFORM),linux/amd64)
	sed -i "" 's/\$$BUILDPLATFORM/linux\/amd64/g' Dockerfile
	docker build -t $(REGISTRY):$(TAG) -f Dockerfile .
	@echo + Building image $(REGISTRY):$(TAG) successfully
else
	@echo + Building multi arch for platform $(TARGETPLATFORM)
	docker buildx build \
		--platform $(TARGETPLATFORM) \
		-t $(MULTI_ARCH_REGISTRY):$(TAG) -f Dockerfile . \
		--push
	@echo + Building image $(MULTI_ARCH_REGISTRY):$(TAG) successfully
endif

.PHONY: push
push: image
ifeq ($(TARGETPLATFORM),linux/amd64)
	docker push $(REGISTRY):$(TAG)
else
	docker push $(MULTI_ARCH_REGISTRY):$(TAG)
endif

.PHONY: ccm-mac, ccm-linux, ccm-win, ccm-arm64
ccm-mac:
	GOARCH=amd64 \
	GOOS=darwin \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	go build -mod vendor -v -o build/bin/cloud-controller-manager \
       -ldflags $(ldflags) cmd/manager/main.go
ccm-linux:
	GOARCH=amd64 \
	GOOS=linux \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	go build -mod vendor -v -o build/bin/cloud-controller-manager.amd64 \
       -ldflags $(ldflags) cmd/manager/main.go
ccm-win:
	GOARCH=amd64 \
	GOOS=windows \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	go build -mod vendor -v -o build/bin/cloud-controller-manager.exe \
       -ldflags $(ldflags) cmd/manager/main.go
ccm-arm64:
	GOARCH=arm64 \
	GOOS=linux \
	CGO_ENABLED=0 \
	GO111MODULE=on \
	go build -mod vendor -v -o build/bin/cloud-controller-manager.arm64 \
       -ldflags $(ldflags) cmd/manager/main.go

.PHONY: check
check: gofmt golint

.PHONY: gofmt
gofmt:
	./hack/verify-gofmt.sh

.PHONY: golint
golint:
	export GO111MODULE=on
	GOLANGCI_LINT_CACHE=/tmp/.cache GOFLAGS="-mod=vendor"
	${GOLANGCI_LINT} run pkg/... -v

unit-test:
	GO111MODULE=on go test -mod vendor -v -race -coverprofile=coverage.txt -covermode=atomic \
		k8s.io/cloud-provider-alibaba-cloud/pkg/...
