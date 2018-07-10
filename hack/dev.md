## Dev

This is a guide for debug cloud-controller-manager in your local intellij idea environment.

### Step 1.

Go to alibaba cloud control [panel](https://cs.console.aliyun.com), create a kubernetes cluster.
Go to [manager infomation] page find your cluster connection configuration, and copy to local. See below.
```
mkdir $HOME/.kube
scp root@$APISERVER:/etc/kubernetes/kube.conf $HOME/.kube/config
```

Note apiserver ip as ```APISERVER_IP```.

Run ```kubectl get po``` to verify. How to install kubectl?

### Step 2. Stop controller

Stop your kubernetes cluster cloud-controller-manager component.

```
    ### rename your cloud-controller-manager image.
    kubectl get ds -n kube-system cloud-controller-manager -o yaml |grep image:|awk -F "image: " '{print $2}'|xargs -I '{}' kubectl set image ds/cloud-controller-manager -n kube-system cloud-controller-manager={}-bak

    ### restart your cloud-controller-manager.
    kubectl get po -n kube-system|grep cloud-con|awk '{print $1}'|xargs -I '{}' kubectl delete po -n kube-system {}
```

### Step 3. Set up your metaserver proxy.

run ```bash hack/setup.sh```

### Step 4. Setup proxy to your kubernetes cluster.

replaced with your own apiserver public loadbalancer ip.
```
### replaced with your own apiserver public loadbalancer ip.
sudo ssh -v -N -L 127.0.0.1:80:127.0.0.1:31977 root@$APISERVER_IP

```

### Step 5. Setup your intellij idea

You need ACCESS_KEY_ID and ACCESS_KEY_SECRET . Find it from alibaba cloud control [panel](https://ak-console.aliyun.com/?spm=5176.2020520152.aliyun_topbar.193.5bd916ddyIgotF#/accesskey).
Create a AccessKey, in case you dont have one.

Edit intellij idea [Run configuration] set up three envs below.
Edit intellij idea [Run configuration] set up run parameters below. pls replace $APISERVER with your own.
```
###  set your IDE runner, With Env, With Parameters.
export ACCESS_KEY_ID
export ACCESS_KEY_SECRET
export METADATA_ENDPOINT=http://127.0.0.1

--kubeconfig=/etc/kubernetes/cloud-controller-manager.conf
--address=127.0.0.1
--allow-untagged-cloud=true
--leader-elect=true
--cloud-provider=alicloud
--allocate-node-cidrs=true
--cluster-cidr=172.16.0.0/16
--use-service-account-credentials=true
--route-reconciliation-period=60s
--v=5
--master=https://$APISERVER_IP:6443
```


### Step 5.

Lets just rock and roll