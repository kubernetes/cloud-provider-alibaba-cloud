VERSION ?= v0.1.0
REGISTRY ?= registry.cn-hangzhou.aliyuncs.com/google-containers

all: clean binary build push

clean:
	rm -f alicloud-controller-manager

binary:
	GOOS=linux go build -o alicloud-controller-manager .

build:
	docker build -t $(REGISTRY)/alicloud-controller-manager:$(VERSION) .

push:
	docker push $(REGISTRY)/alicloud-controller-manager:$(VERSION)

.PHONY: clean binary build push
