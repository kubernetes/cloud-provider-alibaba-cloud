
## Prerequisites
- Version: kubernetes version > 1.7.2 is required.
- CloudNetwork: Only Alibaba Cloud VPC network is supported.


## Deploy out-of-tree CloudProvider in Alibaba Cloud.

### Set up a latest supported Kubernetes Cluster of version v1.10+ with Kubeadm.

Kubeadm is an official installation tool for kubernetes. You could bring up a single master kubernetes cluster by following the instruction in this [page](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/).

1. Install Docker or other CRI runtime: https://kubernetes.io/docs/setup/cri/
2. Install kubeadm, kubelet and kubectl: https://kubernetes.io/docs/setup/independent/install-kubeadm/
3. Update kubelet info with provider id info and restart kubelet: You should provide ```--hostname-override=${REGION_ID}.${INSTANCE_ID} --provider-id=${REGION_ID}.${INSTANCE_ID}``` arguments in all of your kubelet unit file. The format is ```${REGION_ID}.${INSTANCE_ID}```. See [kubelet.service](examples/kubelet.service) for more details.
4. Init kubeadm: Be advised that kubeadm accept a serious of certain parameters to customize your cluster with kubeadm.conf file. If you want to use your own secure ETCD cluster or image repository, you may find the template [kubeadm.conf](examples/kubeadm.conf) or [kubeadm-new.conf for k8s 1.12+](examples/kubeadm-new.conf) is useful. 

Run the command below to initialize a kubernetes cluster.
```$bash
kubeadm init --config kubeadm.conf
```

>> Note:
1. ```cloudProvider: external``` is required to set in kubeadm.conf file for you to deploy alibaba out-of-tree cloudprovider.
2. Set ```imageRepository: registry-vpc.${region}.aliyuncs.com/acs``` is a best practice to enable you the ability to pull image faster in China (for example: cn-hangzhou or cn-hongkong).
3. If you are not sure how to find your ECS instance's ID and region id, try to run these command in your ECS instance:
```bash
$ META_EP=http://100.100.100.200/latest/meta-data
$ echo `curl -s $META_EP/region-id`.`curl -s $META_EP/instance-id`
```


For now, you should have a running kubernetes cluster. Try some example command like ```kubectl get no ```

### Install Alibaba CloudProvider support.

CloudProvider needs certain permissions to access Alibaba Cloud, you will need to create a few RAM policies for your ECS instances or use AccessKeyID&Secret directly.

**RAM role Policy**

[What is the RAM role of an instance](https://www.alibabacloud.com/help/doc-detail/54235.htm)

The sample [master policy](examples/master.policy) is a bit open and can be scaled back depending on the use case. Adjust these based on your needs.

**AccessKeyID & AccessKeySecret**

Or we use Alibaba AccessKeyID&Secret to authorize the CloudProvider. Please make sure that the AccessKeyID has the listed permissions in [master.policy](examples/master.policy)

[How to get AccessKey?](https://usercenter.console.aliyun.com/#/manage/ak)

Then, create ```cloud-config``` configmap in the cluster.

```bash
$ kubectl -n kube-system create configmap cloud-config \
        --from-literal=special.keyid="$ACCESS_KEY_ID" \
        --from-literal=special.keysecret="$ACCESS_KEY_SECRET"
```

**ServiceAccount system:cloud-controller-manager**

CloudProvider use system:cloud-controller-manager service account to authorize Kubernetes cluster with RBAC enabled. So:
1. Certain RBAC roles and bindings must be created. See [cloud-controller-manager.yml](examples/cloud-controller-manager.yml) for details.

2. kubeconfig file must be provider. Save the file below to ```/etc/kubernetes/cloud-controller-manager.conf```. And replace ```$CA_DATA``` with the output of command ```cat /etc/kubernetes/pki/ca.crt|base64 -w 0```. And replace servers with your own apiserver address.

```
kind: Config
contexts:
- context:
    cluster: kubernetes
    user: system:cloud-controller-manager
  name: system:cloud-controller-manager@kubernetes
current-context: system:cloud-controller-manager@kubernetes
users:
- name: system:cloud-controller-manager
  user:
    tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: $CA_DATA
    server: https://192.168.1.76:6443
  name: kubernetes
``` 

**Apply CloudProvider daemonset**

An available cloudprovider daemonset yaml file is being prepared in [cloud-controller-manager.yml](examples/cloud-controller-manager.yml). The only thing you need to do is to replace the ${CLUSTER_CIDR} with your own real cluster cidr. 
And then ``` kubectl apply -f examples/cloud-controller-manager.yml``` to finish the installation. 

>> Note:
1. If you use RAM role policy , please delete env ACCESS_KEY_ID and ACCESS_KEY_SECRET from cloud-controller-manager.yml.


## Try With Simple Example
Once `cloud-controller-manager` is up and running, run a sample nginx deployment:
```bash
$ cat <<EOF >nginx.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx-example
spec:
  replicas: 1
  revisionHistoryLimit: 2
  template:
    metadata:
      labels:
        app: nginx-example
    spec:
      containers:
      - image: nginx:latest
        name: nginx
        ports:
          - containerPort: 80
EOF

$ kubectl create -f nginx.yaml
```

Then create service with type: LoadBalancer:
```bash
$ kubectl expose deployment nginx-example --name=nginx-example --type=LoadBalancer --port=80
$ kubectl get svc
NAME            CLUSTER-IP        EXTERNAL-IP     PORT(S)        AGE
nginx-example   192.168.250.19    106.xx.xx.xxx   80:31205/TCP   5s
```
