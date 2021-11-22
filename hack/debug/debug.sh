#!/bin/bash

ME=$(dirname "$0")
WORKINGDIR=~/.kube

mkdir -p ${WORKINGDIR}

echo "
前置条件：
  1. 手动拷贝一个待测试的客户集群的公网kubeconfig文件，保存为${WORKINGDIR}/config
  2. 通过环境变量设置用户账号的AK、Region、vpc、vswitch信息。DefaultRegion=cn-shenzhen.
     export KEY= SECRET= REGION= VPC= VSW=
  注意：vsw的格式为zoneId:vswId, e.g. cn-hangzhou-b:vsw-bpxxxxxx
"

if [[ ! -f ${WORKINGDIR}/config ]];
then
  echo "kubeconfig file must exist[${WORKINGDIR}/config]. public access is needed"; exit 2
fi

if [[ -z ${KEY} || -z ${SECRET} || -z ${VPC} || -z ${VSW}  ]];
then
  echo "AccessKey & AccessKeySecret & VPC & VSW must be specified.  export KEY= SECRET= REGION= VPC= VSW="; exit 2
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
    "vpcid": "${VPC}",
    "vswitchid": "${VSW}",
    "routeTableIDs": "",
    "accessKeyID": "$(echo -n "${KEY}"|base64)",
    "accessKeySecret": "$(echo -n "${SECRET}"|base64)",
    "disablePublicSLB": false
  }
}
EOF
if [[ "$POD_CIDR" == "" ]];
then
   ROUTE=" --configure-cloud-routes=false "
else
  ROUTE=" --route-reconciliation-period=3m \
      --configure-cloud-routes=true \
      --allocate-node-cidrs=true \
      --cluster-cidr=${POD_CIDR}"
fi
cmd="./cloud-controller-manager \
      --kubeconfig=${WORKINGDIR}/config \
      --cloud-config=${WORKINGDIR}/cloud-config \
      --network="public"\
      --v=3 \
      ${ROUTE} \
      "
echo
echo "通过以下启动命令运行CCM"
echo "${cmd}"
echo