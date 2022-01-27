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

# Simple makefile to build kino quickly and reproducibly in a container
# Only requires docker on the host

# settings
REPO_ROOT:=${CURDIR}
# autodetect host GOOS and GOARCH by default, even if go is not installed
GOOS?=linux
GOARCH?=amd64
REGISTRY?=registry.cn-hangzhou.aliyuncs.com/acs/cloud-controller-manager-amd64
TAG?=$(shell git describe --tags)

# make install will place binaries here
# the default path attempst to mimic go install
INSTALL_DIR?=$(shell hack/util/goinstalldir.sh)

# the output binary name, overridden when cross compiling
KIND_BINARY_NAME?=cloud-controller-manager
# use the official module proxy by default
GOPROXY?=https://goproxy.cn,direct
# default build image
GO_VERSION?=1.16
GO_IMAGE?=golang:$(GO_VERSION)
# docker volume name, used as a go module / build cache
CACHE_VOLUME?=cloud-controller-manager-build-cache

# variables for consistent logic, don't override these
CONTAINER_REPO_DIR=/src/cloud-controller-manager
CONTAINER_OUT_DIR=$(CONTAINER_REPO_DIR)/bin
OUT_DIR=$(REPO_ROOT)/bin
UID:=$(shell id -u)
GID:=$(shell id -g)

# ldflags
VERSION_PKG=k8s.io/cloud-provider-alibaba-cloud/version
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date +%Y-%m-%dT%H:%M:%S%z)
ldflags="-s -w -X $(VERSION_PKG).Version=$(TAG) -X $(VERSION_PKG).GitCommit=${GIT_COMMIT} -X ${VERSION_PKG}.BuildDate=${BUILD_DATE}"

# standard "make" target -> builds
all: build

# creates the cache volume
make-cache:
	@echo + Ensuring build cache volume exists
	docker volume create $(CACHE_VOLUME)

# cleans the cache volume
clean-cache:
	@echo + Removing build cache volume
	docker volume rm $(CACHE_VOLUME)

# creates the output directory
out-dir:
	@echo + Ensuring build output directory exists
	mkdir -p $(OUT_DIR)

# cleans the output directory
clean-output:
	@echo + Removing build output directory
	rm -rf $(OUT_DIR)/

# builds cloud-controller-manager in a container, outputs to $(OUT_DIR)
cloud-controller-manager: make-cache out-dir
	@echo + Building cloud-controller-manager binary
	docker run \
		--rm \
		-v $(CACHE_VOLUME):/go \
		-e GOCACHE=/go/cache \
		-v $(OUT_DIR):/out \
		-v $(REPO_ROOT):$(CONTAINER_REPO_DIR) \
		-w $(CONTAINER_REPO_DIR) \
		-e GO111MODULE=on \
		-e GOPROXY=$(GOPROXY) \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-e HTTP_PROXY=$(HTTP_PROXY) \
		-e HTTPS_PROXY=$(HTTPS_PROXY) \
		-e NO_PROXY=$(NO_PROXY) \
		--user $(UID):$(GID) \
		$(GO_IMAGE) \
		go build -v -o /out/$(KIND_BINARY_NAME) \
		    -ldflags $(ldflags) cmd/manager/main.go
	@echo + Built cloud-controller-manager binary to $(OUT_DIR)/$(KIND_BINARY_NAME)

# alias for building cloud-controller-manager
build: cloud-controller-manager


image: build
	docker build -t $(REGISTRY):$(TAG) -f Dockerfile .

bimage: ccm-linux
	docker build --no-cache -t $(REGISTRY):$(TAG) -f Dockerfile .

push: bimage
	docker push $(REGISTRY):$(TAG)

# use: make install INSTALL_DIR=/usr/local/bin
install: build
	@echo + Copying cloud-controller-manager binary to INSTALL_DIR
	install $(OUT_DIR)/$(KIND_BINARY_NAME) $(INSTALL_DIR)/$(KIND_BINARY_NAME)

gen-deploy:
	@export region=cn-shenzhen version=$(TAG) ; bash deploy/gen-deploy.sh


#-X gitlab.alibaba-inc.com/cos/ros.Template=$(ROS_TPL)
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
# standard cleanup target
clean: clean-cache clean-output

.PHONY: all make-cache clean-cache out-dir clean-output cloud-controller-manager build install clean

check:
	gometalinter --disable-all --skip vendor -E ineffassign -E misspell -d ./...



check: gofmt golangci-lint

gofmt:
	./hack/verify-gofmt.sh

golangci-lint:
	bash -c 'golangci-lint run'

unit-test:
	GO111MODULE=on go test -mod vendor -v -race -coverprofile=coverage.txt -covermode=atomic \
		k8s.io/cloud-provider-alibaba-cloud/pkg/...
