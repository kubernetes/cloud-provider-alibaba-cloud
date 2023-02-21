FROM golang:1.18.8

WORKDIR /cloud-provider-alibaba-cloud

ENV GOPROXY=https://goproxy.cn

COPY cmd /cloud-provider-alibaba-cloud/cmd
COPY version /cloud-provider-alibaba-cloud/version
COPY Makefile /cloud-provider-alibaba-cloud/Makefile
COPY go.mod /cloud-provider-alibaba-cloud/go.mod
COPY go.sum /cloud-provider-alibaba-cloud/go.sum
COPY vendor /cloud-provider-alibaba-cloud/vendor
COPY hack /cloud-provider-alibaba-cloud/hack
COPY pkg /cloud-provider-alibaba-cloud/pkg
COPY test /cloud-provider-alibaba-cloud/test
RUN cd /cloud-provider-alibaba-cloud/test/e2e && go test ./ -c -o alb.test