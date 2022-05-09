
## Prerequisites
- Version: kubernetes version > 1.7.2 is required. If using alb ingress, kubernetes version > 1.19.0 is required
- CloudNetwork: Only Alibaba Cloud VPC network is supported.


## Deploy CloudProvider in Alibaba Cloud 

### Set up a supported Kubernetes Cluster using kubeadm

kubeadm is an official installation tool for kubernetes. You could bring up a single master kubernetes cluster by following the instruction in this [page](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/).

1. Install Docker or other CRI runtime: https://kubernetes.io/docs/setup/cri/
2. Install kubeadm, kubelet and kubectl: https://kubernetes.io/docs/setup/independent/install-kubeadm/
3. **MUST** update kubelet info with provider id info and restart kubelet for each node, e.g.
```bash
if [[ 0 -eq `grep '\--provider-id' /etc/systemd/system/kubelet.service.d/10-kubeadm.conf|wc -l` ]]; then
  META_EP=http://100.100.100.200/latest/meta-data
  provider_id=`curl -s $META_EP/region-id`.`curl -s $META_EP/instance-id`
  sed -i "s/--cloud-provider=external/--cloud-provider=external --hostname-override=${provider_id} --provider-id=${provider_id}/g" /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
  systemctl daemon-reload
  systemctl restart kubelet
fi
```
4. Init kubeadm: Be advised that kubeadm accepts a number of certain parameters to customize your cluster with kubeadm.conf file. If you want to use your own secure ETCD cluster or image repository, you may find the template [kubeadm.conf](examples/kubeadm.conf) or [kubeadm-new.conf for k8s 1.12+](examples/kubeadm-new.conf) is useful. 

Run the command below to initialize a kubernetes cluster.
```$bash
kubeadm init --config kubeadm.conf
```

>> Note:
1. ```cloudProvider: external``` is required to set in kubeadm.conf file for you to deploy alibaba out-of-tree cloudprovider.
2. Set ```imageRepository: registry-vpc.${region}.aliyuncs.com/acs``` is a best practice to enable you the ability to pull image faster in China (for example: cn-hangzhou or cn-hongkong).

### Set up a supported Kubernetes Cluster using rke 

Rancher Kubernetes Engine [rke](https://github.com/rancher/rke) is  another Kubernetes installer.

1. Provision the nodes for your cluster in Alicloud
2. Set up cluster.yml file for deploying the Kubernetes cluster. Make sure to provide hostname_override parameters and insert the hostnames of the cluster nodes using the region-id.instance-id format

```
nodes:
  - address: nnn.nnn.nnn.nnn
    user: root
    role:
      - controlplane
      - etcd
    ssh_key_path: ~/ssh.pem
    internal_address: 172.16.1.29
    hostname_override: "cn-hangzhou.i-bp18j8zzajt93rztiw9g" <- override hostname
  
  - address: nnn.nnn.nnn.nnn
    user: root
    role:
      - worker
    ssh_key_path: ~/ssh.pem
    internal_address: 172.16.1.19
    hostname_override: "cn-hangzhou.i-bp109r2aiuf935xxi4po" <- override hostname
    labels:
      rke.cattle.io/external: nnn.nnn.nnn.nnn
  - address: nnn.nnn.nnn.nnn
    user: root
    role:
      - worker
    ssh_key_path: ~/ssh.pem
    internal_address: 172.16.1.21
    hostname_override: "cn-hangzhou.i-bp16uimj7fl6ze8q5rf3" <- override hostname
    labels:
      rke.cattle.io/external: nnn.nnn.nnn.nnn
  
addon_job_timeout: 90
services:
  kubelet:
    extra_args:
      node-status-update-frequency: 4s
      volume-plugin-dir: /usr/libexec/kubernetes/kubelet-plugins/volume/exec
    extra_binds:
      - /usr/libexec/kubernetes/kubelet-plugins/volume/exec:/usr/libexec/kubernetes/kubelet-plugins/volume/exec
  kube-api:
    service_node_port_range: 10000-12500
    extra_args:
      default-not-ready-toleration-seconds: 30
      default-unreachable-toleration-seconds: 30
  kube-controller:
    extra_args:
      node-monitor-period: 2s
      node-monitor-grace-period: 16s
      pod-eviction-timeout: 30s
    addon_job_timeout: 90
```

3. Provision the cluster using rke

4. Currently rke does not support setting the providerID as a configuration option.
   Once the cluster has been provisioned update the nodes and make sure providerID is set to REGION.NODEID
   
   ```
   kubectl patch node ${NODE_NAME} -p '{"spec":{"providerID": "${NODE_NAME}"}}'
   ```
   For example 
   ```
   kubectl patch node cn-hangzhou.i-bp16uimj7fl6ze8q5rf3 -p '{"spec":{"providerID": "cn-hangzhou.i-bp16uimj7fl6ze8q5rf3"}}'
   ```

### Install Alibaba CloudProvider support.

CloudProvider needs certain permissions to access Alibaba Cloud, you will need to create a few RAM policies for your ECS instances or use AccessKeyID and AccessKeySecret directly.

**RAM role Policy**

[What is the RAM role of an instance](https://www.alibabacloud.com/help/doc-detail/54235.htm)

The sample [master policy](examples/master.policy) is a bit open and can be scaled back depending on the use case. Adjust these based on your needs.

**AccessKeyID and AccessKeySecret**

Or we use Alibaba AccessKeyID AccessKeySecret to authorize the CloudProvider. Please make sure that the AccessKeyID has the listed permissions in [master.policy](examples/master.policy)

[How to get AccessKey?](https://usercenter.console.aliyun.com/#/manage/ak)

Then, create ```cloud-config``` configmap in the cluster.

```bash
# base64 AccessKey & AccessKeySecret
$ echo -n "$AccessKeyID" |base64
$ echo -n "$AcceessKeySecret"|base64

$ cat <<EOF >cloud-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloud-config
  namespace: kube-system
data:
  cloud-config.conf: |-
    {
        "Global": {
            "accessKeyID": "$your-AccessKeyID-base64",
            "accessKeySecret": "$your-AccessKeySecret-base64"
        }
    }
EOF

$ kubectl create -f cloud-config.yaml
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


## Try With Simple Example
Once `cloud-controller-manager` is up and running, run a sample nginx deployment:
```bash
$ kubectl create deployment nginx-example --image=nginx
```

Then create service with type: LoadBalancer:
```bash
$ kubectl expose deployment nginx-example --name=nginx-example-svc --type=LoadBalancer --port=80
$ kubectl get svc
NAME                TYPE           CLUSTER-IP    EXTERNAL-IP     PORT(S)        AGE
nginx-example-svc   LoadBalancer   10.96.38.24   10.x.x.x        80:30536/TCP   38s
```

## Try With Simple ALB Ingress Example
run a sample ingress: [usage-alb.md](usage-alb.md)