## Requirements

### Version
Kubernetes version 1.7.2 or higher is required to get a stable running.

### AliCloud ECS
Only VPC network is supported.

## Getting started
To deploy cloud-controller-manager in kubernetes cluster, we need to do a few things:

- Get an `cloud-controller-manager` image.
- Prepare your kubernetes cluster with some requirements.
- Prepare and deploy `cloud-controller-manager`.
- Try it!

### Get an `cloud-controller-manager` image

You can either get an image from official release by image name `registry.cn-hangzhou.aliyuncs.com/acs/cloud-controller-manager:<RELEASE_VERSION>`

Or build it from source which require docker has been installed:

    ```bash
    # for example. export REGISTRY=registry.cn-hangzhou.aliyuncs.com/acs
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

By default, the kubelet will name nodes based on the node's hostname. But in `cloud-controller-manager`, we use `<REGION_ID>.<ECS_ID>` format to build a unique node id to identity one node. In order to elimite these difference, we need to set extra flags `--hostname_override` and `--provider-id` with format `<REGION_ID>.<ECS_ID>` on kubelet.

If you are not sure how to find your ECS instance's ID and region id, try to run these command in your ECS instance:

    ```bash
    $ META_EP=http://100.100.100.200/latest/meta-data
    $ echo `curl -s $META_EP/region-id`.`curl -s $META_EP/instance-id`
    ```

### Prepare and deploy `cloud-controller-manager`

1. Prepare AliCloud access key id and secret

```
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

1. Prepare `cloud-controller-manager` daemonset yaml

Mare sure container image, `--cluster-cidr` field match what your needs. replace image with your version.

```
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  labels:
    app: cloud-controller-manager
    tier: control-plane
  name: cloud-controller-manager
  namespace: kube-system
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: cloud-controller-manager
      tier: control-plane
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: cloud-controller-manager
        tier: control-plane
    spec:
      containers:
      - command:
        - /cloud-controller-manager
        - --kubeconfig=/etc/kubernetes/cloud-controller-manager.conf
        - --address=127.0.0.1
        - --leader-elect=true
        - --cloud-provider=alicloud
        - --allocate-node-cidrs=true
        - --allow-untagged-cloud=true
        # set this to what you set to controller-manager or kube-proxy
        - --cluster-cidr=172.20.0.0/16
        - --use-service-account-credentials=true
        - --route-reconciliation-period=30s
        - --v=5
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
        image: registry-vpc.cn-hangzhou.aliyuncs.com/acs/cloud-controller-manager-amd64:v1.9.3
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 8
          httpGet:
            host: 127.0.0.1
            path: /healthz
            port: 10252
            scheme: HTTP
          initialDelaySeconds: 15
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 15
        name: cloud-controller-manager
        resources:
          requests:
            cpu: 200m
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/kubernetes/
          name: k8s
          readOnly: true
        - mountPath: /etc/ssl/certs
          name: certs
        - mountPath: /etc/pki
          name: pki
      dnsPolicy: ClusterFirst
      hostNetwork: true
      nodeSelector:
        node-role.kubernetes.io/master: ""
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: cloud-controller-manager
      serviceAccountName: cloud-controller-manager
      terminationGracePeriodSeconds: 30
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      - effect: NoSchedule
        key: node.cloudprovider.kubernetes.io/uninitialized
        operator: Exists
      volumes:
      - hostPath:
          path: /etc/kubernetes
          type: ""
        name: k8s
      - hostPath:
          path: /etc/ssl/certs
          type: ""
        name: certs
      - hostPath:
          path: /etc/pki
          type: ""
        name: pki
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate

```
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