FROM ubuntu:16.04

RUN apt-get update -y && apt-get install -y ca-certificates

ADD alicloud-controller-manager /

CMD ["/alicloud-controller-manager"]
