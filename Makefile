
##############################################################################################################

.PHONY: test e2e-test cover gofmt gofmt-fix  clean

# Registry used for publishing images
REGISTRY?=registry.cn-hangzhou.aliyuncs.com/google-containers/cloud-controller-manager-amd64

# Default tag and architecture. Can be overridden
TAG?=$(shell git describe --tags)
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
	  -ldflags "-X k8s.io/cloud-provider-alicloud/version.Version=$(TAG)"

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

pre-requisite:
	@echo "Warning: Tag your branch before make. or makefile can not autodetect image tag."

test:
	go test -v k8s.io/cloud-provider-alicloud/alicloud-controller-manager/cloudprovider/alicloud

image: cloud-controller-manager-$(ARCH)
	docker build -f build/Dockerfile -t $(REGISTRY):$(TAG) ./build/

docker-push:
	docker push $(REGISTRY):$(TAG)

# amd64 gets an image with the suffix too (i.e. it's the default)
ifeq ($(ARCH),amd64)
	docker push $(REGISTRY):$(TAG)
endif

docker-build: cloud-controller-manager-$(ARCH)

## Build an architecture specific cloud-controller-manager binary
cloud-controller-manager-$(ARCH): pre-requisite
	# Build for other platforms with ARCH=$$ARCH make build
	# valid values for $$ARCH are [amd64 arm arm64 ppc64le]

	docker run -e CC=$(CC) -e GOARM=$(GOARM) -e GOARCH=$(ARCH) \
		-v $(SOURCE):/go/src/k8s.io/cloud-provider-alicloud \
		-v $(SOURCE)/build:/go/src/k8s.io/cloud-provider-alicloud/build \
		registry.cn-hangzhou.aliyuncs.com/google-containers/kube-cross:$(KUBE_CROSS_TAG) /bin/bash -c '\
		cd /go/src/k8s.io/cloud-provider-alicloud && \
		CGO_ENABLED=1 make -e cloud-controller-manager && \
		mv cloud-controller-manager build/cloud-controller-manager-$(ARCH) && \
		file build/cloud-controller-manager-$(ARCH)'
