FROM ubuntu:16.04

# install wget to download golang source code
# install git
RUN apt-get update \
    && apt-get install -y \
    wget \ 
    git \
    make \
    gcc \
    vim \
    tree \
    && apt-get clean

# set go version this image use
ENV GO_VERSION=1.16
ENV ARCH=amd64

# install golang which version is GO_VERSION
RUN wget --quiet https://storage.googleapis.com/golang/go${GO_VERSION}.linux-${ARCH}.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz \
    && rm go${GO_VERSION}.linux-${ARCH}.tar.gz

# create GOPATH
RUN mkdir /go
ENV GOPATH=/go

# set go binary path to local $PATH
# go binary path is /usr/local/go/bin
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV GONOPROXY "gitlab.alibaba-inc.com"
ENV GONOSUMDB "gitlab.alibaba-inc.com"
ENV GOPRIVATE "gitlab.alibaba-inc.com"
#ENV GOPROXY "https://goproxy.cn,direct"
ENV GO111MODULE off

COPY . /go/src/gitlab.alibaba-inc.com/cloud-native-cicd/chorus-e2e
WORKDIR /go/src/gitlab.alibaba-inc.com/cloud-native-cicd/chorus-e2e
RUN go get -u github.com/onsi/ginkgo/ginkgo
RUN cd /go/src/gitlab.alibaba-inc.com/cloud-native-cicd/chorus-e2e && ./hack/generate_bindata.sh && ginkgo build ./
