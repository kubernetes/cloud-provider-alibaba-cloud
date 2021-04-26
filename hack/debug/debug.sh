#!/bin/bash

ME=$(dirname "$0")
WORKINGDIR=~/.kube

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
    "accessKeyID": "$(echo -n "${KEY}"|base64)",
    "accessKeySecret": "$(echo -n "${SECRET}"|base64)",
    "disablePublicSLB": false
  }
}
EOF
if [[ "$POD_CIDR" == "" ]];
then
   ROUTE=" --configure-cloud-routes=false \
      --allocate-node-cidrs=false"
else
  ROUTE=" --route-reconciliation-period=30s \
      --configure-cloud-routes=true \
      --allocate-node-cidrs=true \
      --cluster-cidr=${POD_CIDR}"
fi
cmd="./cloud-controller-manager \
      --kubeconfig=${WORKINGDIR}/config \
      --cloud-config=${WORKINGDIR}/cloud-config \
      --enable-leader-select=true \
      --loglevel=3 \
      ${ROUTE}
      "
echo
echo "通过以下启动命令运行CCM"
echo "${cmd}"
echo