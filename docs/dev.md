
Developers may want to build the image from scratch. We provide a simple command ```make image``` to accomplish this. 
Be advise that build it from source requires docker to be installed.
A valid tag is required to build your image. Tag with ```git tag v1.9.3```

## Build `cloud-controller-manager` image

```bash
# for example. export REGISTRY=registry.cn-hangzhou.aliyuncs.com/acs
$ export REGISTRY=<YOUR_REGISTRY_NAME>
# This will build cloud-controller-manager from source code and build an docker image from binary and push to your specified registry.
# You can also use `make binary && make build` if you don't want push this image to your registry.
$ make image
$ docker images |grep cloud-controller-manager
```


## Testing

### UnitTest

See [Testing UnitTest](https://github.com/kubernetes/cloud-provider-alibaba-cloud/tree/master/docs/testing.md)

### E2ETest

See [Testing E2E](https://github.com/kubernetes/cloud-provider-alibaba-cloud/tree/master/cmd/e2e/README.md)
