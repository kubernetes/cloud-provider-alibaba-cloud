
##############################################################################################################

.PHONY: test e2e-test cover gofmt gofmt-fix  clean

# Registry used for publishing images
REGISTRY?=registry.cn-hangzhou.aliyuncs.com/google-containers

# Default tag and architecture. Can be overridden
TAG?=$(shell git describe --tags --dirty)
ARCH?=amd64

# Set the (cross) compiler to use for different architectures
ifeq ($(ARCH),amd64)
	LIB_DIR=x86_64-linux-gnu
	CC=gcc
endif

ifeq ($(LOCAL),)
	SOURCE=$(shell echo ${PWD})
else
    SOURCE=$(LOCAL)
endif
GOARM=6
KUBE_CROSS_TAG=v1.8.1-1
IPTABLES_VERSION=1.4.21

cloud-controller-manager: $(shell find . -type f  -name '*.go')
	go build -o cloud-controller-manager \
	  -ldflags "-X github.com/AliyunContainerService/alicloud-controller-manager/version.Version=$(TAG)"

# Throw an error if gofmt finds problems.
# "read" will return a failure return code if there is no output. This is inverted wth the "!"
gofmt:
	bash -c '! gofmt -d $(PACKAGES) 2>&1 | read'

gofmt-fix:
	gofmt -w $(PACKAGES)


clean:
	rm -f cloud-controller-manager*
	rm -f dist/iptables*
	rm -f dist/*.aci
	rm -f dist/*.docker
	rm -f dist/*.tar.gz

## Create a docker image on disk for a specific arch and tag
cloud-controller-manager-$(TAG)-$(ARCH).docker: cloud-controller-manager-$(ARCH) dist/iptables-$(ARCH) dist/libpthread.so.0-$(ARCH)
	docker build -f Dockerfile.$(ARCH) -t $(REGISTRY):$(TAG)-$(ARCH) .
	docker save -o cloud-controller-manager-$(TAG)-$(ARCH).docker $(REGISTRY):$(TAG)-$(ARCH)

# amd64 gets an image with the suffix too (i.e. it's the default)
ifeq ($(ARCH),amd64)
	docker build -f Dockerfile -t $(REGISTRY):$(TAG) ./build/
endif
image:
	docker build -f Dockerfile -t $(REGISTRY):$(TAG) .
docker-push: cloud-controller-manager-$(TAG)-$(ARCH).docker
	docker push $(REGISTRY):$(TAG)-$(ARCH)

# amd64 gets an image with the suffix too (i.e. it's the default)
ifeq ($(ARCH),amd64)
	docker push $(REGISTRY):$(TAG)
endif

docker-build: cloud-controller-manager-$(ARCH)

## Build an architecture specific cloud-controller-manager binary
cloud-controller-manager-$(ARCH):
	# Build for other platforms with ARCH=$$ARCH make build
	# valid values for $$ARCH are [amd64 arm arm64 ppc64le]

	docker run -e CC=$(CC) -e GOARM=$(GOARM) -e GOARCH=$(ARCH) \
		-v $(SOURCE):/go/src/github.com/AliyunContainerService/alicloud-controller-manager \
		-v $(SOURCE)/build:/go/src/github.com/AliyunContainerService/alicloud-controller-manager/build \
		registry.cn-hangzhou.aliyuncs.com/google-containers/kube-cross:$(KUBE_CROSS_TAG) /bin/bash -c '\
		cd /go/src/github.com/AliyunContainerService/alicloud-controller-manager && \
		CGO_ENABLED=1 make -e cloud-controller-manager && \
		mv cloud-controller-manager build/cloud-controller-manager-$(ARCH) && \
		file cloud-controller-manager-$(ARCH)'
