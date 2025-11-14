FROM --platform=$BUILDPLATFORM golang:1.24 as builder
ADD . /build
WORKDIR /build/
ARG TARGETOS
ARG TARGETARCH
RUN make cloud-controller-manager

FROM registry.cn-hangzhou.aliyuncs.com/acs/alpine:3.18-update
RUN apk add --no-cache --update bash ca-certificates
COPY --from=builder /build/bin/cloud-controller-manager /cloud-controller-manager

ENTRYPOINT  ["/cloud-controller-manager"]
