##############################################################################################################

.PHONY: test e2e-test cover gofmt gofmt-fix  clean cloud-controller-manager

# Registry used for publishing images
REGISTRY?=registry.cn-hangzhou.aliyuncs.com/google-containers/cloud-controller-manager-amd64

# Default tag and architecture. Can be overridden
TAG?=$(shell git describe --tags)
ARCH?=amd64

CGO_ENABLED=1
# Set the (cross) compiler to use for different architectures
ifeq ($(ARCH),amd64)
	LIB_DIR=x86_64-linux-gnu
	CC=gcc
	CGO_ENABLED=0
endif

ifeq ($(LOCAL),)
	SOURCE=$(shell echo ${PWD})
else
	SOURCE=$(LOCAL)
endif
GOARM=6
KUBE_CROSS_TAG=v1.9.3-1
IPTABLES_VERSION=1.4.21

cloud-controller-manager: $(shell find . -type f  -name '*.go')
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloud-controller-manager-alicloud -ldflags "-X k8s.io/cloud-provider-alibaba-cloud/version.Version=$(TAG)" \
		cmd/cloudprovider/cloudprovider-alibaba-cloud.go


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

unit-test:
	pwd && go test -race -coverprofile=coverage.txt -covermode=atomic \
		github.com/kubernetes/cloud-provider-alibaba-cloud/cloud-controller-manager \
		github.com/kubernetes/cloud-provider-alibaba-cloud/cloud-controller-manager/controller/route

check:
	gometalinter --disable-all --skip vendor -E ineffassign -E misspell -d ./...

test:
	#go test -v k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/alicloud
	docker run -e CC=$(CC) -e GOARM=$(GOARM) -e GOARCH=$(ARCH) -e CGO_ENABLED=$(CGO_ENABLED) \
		-v $(SOURCE):/go/src/k8s.io/cloud-provider-alibaba-cloud \
		-v $(SOURCE)/build:/go/src/k8s.io/cloud-provider-alibaba-cloud/build \
		registry.cn-hangzhou.aliyuncs.com/google-containers/kube-cross:$(KUBE_CROSS_TAG) /bin/bash -c '\
		cd /go/src/k8s.io/cloud-provider-alibaba-cloud && \
		go test -v k8s.io/cloud-provider-alibaba-cloud/cloud-controller-manager/alicloud'

image: cloud-controller-manager-$(ARCH)
	docker build -t $(REGISTRY):$(TAG) .

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

	docker run -e CC=$(CC) -e GOARM=$(GOARM) -e GOARCH=$(ARCH) -e CGO_ENABLED=$(CGO_ENABLED) \
		-v $(SOURCE):/go/src/k8s.io/cloud-provider-alibaba-cloud \
		-v $(SOURCE)/build:/go/src/k8s.io/cloud-provider-alibaba-cloud/build \
		registry.cn-hangzhou.aliyuncs.com/google-containers/kube-cross:$(KUBE_CROSS_TAG) /bin/bash -c '\
		cd /go/src/k8s.io/cloud-provider-alibaba-cloud && \
		TAG=$(TAG) make -e cloud-controller-manager && \
		mv cloud-controller-manager-alicloud build/cloud-controller-manager-$(ARCH) && \
		file build/cloud-controller-manager-$(ARCH)'
