#!/bin/bash
set -e -x
if [ -z $APISERVER_IP ];then
	APISERVER_IP=47.110.78.7
fi
### rename your cloud-controller-manager image.
kubectl get ds -n kube-system cloud-controller-manager -o yaml |grep image:|awk -F "image: " '{print $2}'|xargs -I '{}' kubectl set image ds/cloud-controller-manager -n kube-system cloud-controller-manager={}-bak

### restart your cloud-controller-manager.
kubectl -n kube-system delete po -lapp=cloud-controller-manager

setup_localproxy()
{
	sudo mkdir -p /etc/kubernetes/ /var/run/secrets/kubernetes.io/serviceaccount/
	echo "1. check to see whether metaserver proxy is running..."
	cnt=$(kubectl get deploy 2>&1|grep metaserver|wc -l)
	if [[ "$cnt" == *"0"* ]];then
	    kubectl apply -f hack/metaserver-deploy.yaml
	    kubectl apply -f hack/metaserver-svc.yaml
	    echo "    run new metaserver successfully."
	fi

	echo "2. copy cloud-controller-manager.conf to /etc/kubernetes/cloud-controller-manager.conf"
	scp root@$APISERVER_IP:/etc/kubernetes/cloud-controller-manager.conf cloud-controller-manager.conf
	SERVER=$(grep "server: https:/" cloud-controller-manager.conf)
	sed -i '' "s@$SERVER@    server: https://$APISERVER_IP:6443@" cloud-controller-manager.conf
	NAME=$(grep "name: c" cloud-controller-manager.conf)
	if [ -n "$NAME" ];then
	    sed -i '' "s@$NAME@  name: kubernetes@" cloud-controller-manager.conf
	fi
	sudo cp cloud-controller-manager.conf /etc/kubernetes/cloud-controller-manager.conf
#	cat /etc/kubernetes/cloud-controller-manager.conf
	sudo rm -rf cloud-controller-manager.conf
	echo "3. prepare token file /var/run/secrets/kubernetes.io/serviceaccount/token"
	token=$(kubectl get secret -n kube-system |grep cloud-controller-manager |awk '{print $1}'|xargs -I '{}' kubectl get secret -n kube-system {} -o yaml|grep token:|awk -F "token: " '{print $2}'|base64 --decode |tr -d '\n')
	echo -n $token >token
	sudo mv token /var/run/secrets/kubernetes.io/serviceaccount/token
}


setup_localproxy