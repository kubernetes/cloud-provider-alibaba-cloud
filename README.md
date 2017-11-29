# Kubernetes Cloud Controller Manager for Alibaba Cloud

`cloud-controller-manager` is the external Kubernetes cloud controller manager implementation for AliCloud(Alibaba Cloud). Running `cloud-controller-manager` allows you build your kubernetes clusters leverage on many cloud services on AliCloud. You can read more about Kubernetes cloud controller manager [here](https://kubernetes.io/docs/tasks/administer-cluster/running-cloud-controller/).

**WARNING:** This project is still work in progress, be careful using it in production environment.

## Requirements

### Version
Kubernetes version 1.7.2 or higher is required to get a stable running.

### AliCloud ECS
Only VPC network is supported.

## Getting started
To deploy cloud-controller-manager in kubernetes cluster, we need to a few things:

- Get an `cloud-controller-manager` image.
- Prepare your kubernetes cluster with some requirements.
- Prepare and deploy `cloud-controller-manager`.
- Try it!

### Get an `cloud-controller-manager` image

You can either get an image from official release by image name `registry.cn-hangzhou.aliyuncs.com/google-containers/cloud-controller-manager:<RELEASE_VERSION>`

Or build it from source which require docker has been installed:

    ```bash
    # for example. export REGISTRY=registry.cn-hangzhou.aliyuncs.com/google-containers
    $ export REGISTRY=<YOUR_REGISTRY_NAME>
    # This will build cloud-controller-manager from source code and build an docker image from binary and push to your specified registry.
    # You can also use `make binary && make build` if you don't want push this image to your registry.
    $ make image
    $ docker images |grep cloud-controller-manager
    ```

### Prepare your kubernetes cluster with some requirements

#### --cloud-provider=external

In order to external cloud provider feature, we need to deploy or reconfigure `kube-apiserver`/`kube-controller-manager`/`kubelet` component with extra flag `--cloud-provider=external`, which means cloud provider functionality will hand to out of tree external cloud provider, here we use `cloud-controller-manager`.

How and where to set this flag depends on how you deploy your cluster, we will give a detail `kubeadm` way to deploy cluster with `cloud-controller-manager` later.

#### hostname and provider id

By default, the kubelet will name nodes based on the node's hostname. But in `cloud-controller-manager`, we use `<REGION_ID>.<ECS_ID>` format to build a unique node id to identity one node. In order to elimite these difference, we need to set extra flags `--hostname_override` and `--provider-id` to `<REGION_ID>.<ECS_ID>`.

If you are not sure how to find your ECS instance's ID and region id, try to run these command in your ECS instance:

    ```bash
    $ META_EP=http://100.100.100.200/latest/meta-data
    $ echo `curl -s $META_EP/region-id`.`curl -s $META_EP/instance-id`
    ```

### Prepare and deploy `cloud-controller-manager`

1. Prepare AliCloud access key id and secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloud-config
  namespace: kube-system
data:
  # insert your base64 encoded AliCloud access id and key here, ensure there's no trailing newline:
  # to base64 encode your token run:
  #      echo -n "abc123abc123doaccesstoken" | base64
  access-key-id: "<ACCESS_KEY_ID>"
  access-key-secret: "<ACCESS_KEY_SECRET>"
```

2. Prepare `cloud-controller-manager` deployment yaml
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: cloud-controller-manager
  namespace: kube-system
spec:
  replicas: 1
  revisionHistoryLimit: 2
  template:
    metadata:
      labels:
        app: cloud-controller-manager
    spec:
      dnsPolicy: Default
      tolerations:
        # this taint is set by all kubelets running `--cloud-provider=external`
        - key: "node.cloudprovider.kubernetes.io/uninitialized"
          value: "true"
          effect: "NoSchedule"
      containers:
      - image: registry.cn-hangzhou.aliyuncs.com/google-containers/cloud-controller-manager:v1.8.1
        name: cloud-controller-manager
        command:
          - /cloud-controller-manager
          # set leader-elect=true if you have more that one replicas
          - --leader-elect=false
          - --allocate-node-cidrs=true
          # set this to what you set to controller-manager or kube-proxy
          - --cluster-cidr=192.168.0.0/20
          # if you want to use a secure endpoint or deploy in a kubeadm deployed cluster, you need to use a kubeconfig instead.
          - --master=<YOUR_MASTER_INSECURE_ENDPOINT>
        env:
          - name: ACCESS_KEY_ID
            valueFrom:
              secretKeyRef:
                name: cloud-config
                key: access-key-id
          - name: ACCESS_KEY_SECRET
            valueFrom:
              secretKeyRef:
                name: cloud-config
                key: access-key-secret
```
Mare sure container image, `--cluster-cidr` and `--master` field match your needs.

3. Deploy `cloud-controller-manager`
```bash
$ kubectl create -f cloud-controller-manager.yaml
```

### Try it!
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
$ kubectl expose deployment nginx-example --name=nginx-example --type=LoadBalancer
$ kubectl get svc
NAME            CLUSTER-IP        EXTERNAL-IP     PORT(S)        AGE
nginx-example   192.168.250.19    106.xx.xx.xxx   80:31205/TCP   5s
```
