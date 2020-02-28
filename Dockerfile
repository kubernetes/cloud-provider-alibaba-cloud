#############      builder       #############
FROM golang:1.11.5-alpine3.8 AS builder

ARG tag
WORKDIR /go/src/k8s.io/cloud-provider-alibaba-cloud
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install \
    -ldflags "-X k8s.io/cloud-provider-alibaba-cloud/version.Version=$tag" \
    ./...

############# cloud controller manager #############
FROM alpine:3.8

RUN apk add --no-cache --update ca-certificates

COPY --from=builder /go/bin/cloudprovider /cloud-controller-manager

ENTRYPOINT ["/cloud-controller-manager"]
