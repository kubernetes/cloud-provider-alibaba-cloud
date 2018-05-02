#!/bin/bash
set -e
if [ -z $APISERVER_IP ];then
	APISERVER_IP=47.97.236.85
fi

setup_localproxy()
{
	mkdir -p /etc/kubernetes/ /var/run/secrets/kubernetes.io/serviceaccount/
	echo "1. check to see whether metaserver proxy is running..."
	cnt=$(kubectl get deploy metaserver|grep metaserver|wc -l)
	if [ "$cnt" == "0" ];then
	    kubectl run  metaserver --image=registry.cn-hangzhou.aliyuncs.com/spacexnice/nginx-net:latest
	    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  labels:
    run: metaserver
  name: metaserver
  namespace: default
spec:
  ports:
  - nodePort: 31977
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: metaserver
  type: NodePort
EOF
	    echo "    run new metaserver successfully."
	fi

	echo "2. copy cloud-controller-manager.conf to /etc/kubernetes/cloud-controller-manager.conf"
	sudo scp -i ~/.ecs/spacex.spanner.pem root@$APISERVER_IP:/etc/kubernetes/cloud-controller-manager.conf cloud-controller-manager.conf
	SERVER=$(grep "server: https:/" /etc/kubernetes/cloud-controller-manager.conf)

	sed -i '' "s@$SERVER@    server: https://$APISERVER_IP:6443@" cloud-controller-manager.conf
	sudo cp cloud-controller-manager.conf /etc/kubernetes/cloud-controller-manager.conf
#	cat /etc/kubernetes/cloud-controller-manager.conf
	rm cloud-controller-manager.conf
	echo "3. prepare token file /var/run/secrets/kubernetes.io/serviceaccount/token"
	token=$(kubectl get secret -n kube-system |grep cloud-controller-manager |awk '{print $1}'|xargs -I '{}' kubectl get secret -n kube-system {} -o yaml|grep token:|awk -F "token: " '{print $2}'|base64 --decode |tr -d '\n')
	sudo echo -n $token >/var/run/secrets/kubernetes.io/serviceaccount/token
}


setup_localproxy