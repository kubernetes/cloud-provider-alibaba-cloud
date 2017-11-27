FROM ubuntu:16.04
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list
RUN apt-get update -y && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

ADD cloud-controller-manager-amd64 /cloud-controller-manager

CMD ["/cloud-controller-manager"]
