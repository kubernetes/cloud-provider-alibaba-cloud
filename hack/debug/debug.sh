#!/bin/bash

ME=$(dirname "$0")
WORKINGDIR=~/.nlc

mkdir -p ${WORKINGDIR}

echo "
前置条件：
  1. 手动拷贝一个待测试的客户集群的公网kubeconfig文件，保存为${WORKINGDIR}/config
  2. 通过环境变量设置用户账号的AK及Region信息。DefaultRegion=cn-shenzhen.
     export KEY= SECRET= REGION=
"

if [[ ! -f ${WORKINGDIR}/config ]];
then
  echo "kubeconfig file must exist[${WORKINGDIR}/config]. public access is needed"; exit 2
fi

if [[ -z ${KEY} || -z ${SECRET} ]];
then
  echo "AccessKey & AccessKeySecret must be specified. export KEY= SECRET= REGION="; exit 2
fi

if [[ -z $REGION ]];
then
  REGION=cn-shenzhen
fi
# 1. user cluster kubeconfig: config

# 2. cloud config
cat > ${WORKINGDIR}/cloud-config << EOF
{
  "Global": {
    "region": "${REGION}",
    "routeTableIDs": "",
    "accessKey": "$(echo -n "${KEY}"|base64)",
    "accessSecret": $(echo -n "${SECRET}"|base64)",
    "disablePublicSLB": false
  }
}
EOF

cmd="./cloud-controller-manager \
      --kubeconfig=${WORKINGDIR}/config \
      --config=${WORKINGDIR}/cloud-config "
echo
echo "通过以下启动命令运行NLC"
echo "${cmd}"
echo