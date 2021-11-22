FROM registry.cn-hangzhou.aliyuncs.com/acs/alpine:3.13-update

# Do not use docker multiple stage build until we
# figure a way out how to solve build cache problem under 'go mod'.
#

RUN apk add --no-cache --update bash ca-certificates

COPY bin/cloud-controller-manager /cloud-controller-manager

ENTRYPOINT  ["/cloud-controller-manager"]