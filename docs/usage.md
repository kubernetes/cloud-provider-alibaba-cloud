# Alibaba Cloud Provider

## Alibaba Cloud Provider introduction

**CloudProvider** provides the Cloud Provider interface implementation as an out-of-tree cloud-controller-manager. It allows Kubernetes clusters to leverage the infrastructure services of Alibaba Cloud .
It is original open sourced project is [https://github.com/AliyunContainerService/alicloud-controller-manager](https://github.com/AliyunContainerService/alicloud-controller-manager)

[See ReleaseNotes](https://yq.aliyun.com/articles/608575)

**Basic usage** cloudprovider use service annotation to control service creation behavior. Here is a basic annotation example:
```
apiVersion: v1
kind: Service
metadata:
  annotations:
    # here is your annotation, example
    service.beta.kubernetes.io/alicloud-loadbalancer-id: lb-bp1hfycf39bbeb019pg7m
  name: nginx
  namespace: default
spec:
  ports:
  - name: web
    port: 443
    protocol: TCP
    targetPort: 443
  type: LoadBalancer
```

\>> **Note：**    

- CloudProvider would not deal with your LoadBalancer(which was provided by user) listener by default if your cloud-controller-manager version is great equal then v1.9.3. User need to config their listener by themselves or using ```service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners: "true"``` to force overwrite listeners. 

Using the following command to find the version of your cloud-controller-manager 

```
root@master # kubectl get po -n kube-system -o yaml|grep image:|grep cloud-con|uniq

image: registry-vpc.cn-....-controller-manager-amd64:v1.9.3
```

- Some features might not be usable until you upgrade your cloud-controller-manager to the latest version. See[manaully upgrade CloudProvider](https://yq.aliyun.com/articles/608563?spm=a2c4e.11153940.blogrightarea608575.9.57ed1279saZghW)。

## How to create service with Type=LoadBalancer

#### pre-requirement。

- An available ACS kubernetes cluster。[See](https://help.aliyun.com/document_detail/53752.html?spm=a2c4g.11186623.6.567.VslAYT)
- How to connect to your kubernetes cluster with kubectl。[See](https://help.aliyun.com/document_detail/53755.html?spm=a2c4g.11186623.6.572.CgrxgR)
- Create an nginx deployment。[See](https://help.aliyun.com/document_detail/53768.html?spm=a2c4g.11186623.6.586.RsKYlW) , The example below is based on then nginx deployment。

\>> **Note** 

- Save the yaml template to svc.1.yaml ， and then use ```kubectl apply -f svc.1.yaml``` to create your service.

#### 1. Create a public LoadBalancer.

```
apiVersion: v1
kind: Service
metadata:
  name: nginx
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

#### 2. Create a private LoadBalancer.

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-address-type: "intranet"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

#### 3. Create a LoadBalancer with HTTP listener.

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port: "http:80"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

#### 4. Create a LoadBalancer with HTTPS listener.

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port: "https:443"
    service.beta.kubernetes.io/alicloud-loadbalancer-cert-id: ${YOUR_CERT_ID}
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    run: nginx
  type: LoadBalancer
```

\>> **Note：** 

- You need a certificate ID to create an https LoadBalancer. Please heading to the Aliyun Console to create one.

#### 5. Restrict the bandwidth of the LoadBalancer

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth: "100"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    run: nginx
  type: LoadBalancer
```

\>> **Note：** 

- Only restrict the bandwidth of the LoadBalancer. And all listeners share the same bandwidth. See [Bandwidth](https://help.aliyun.com/document_detail/85930.html?spm=a2c4g.11186623.6.640.iPgsrU)

#### 6. Create a specified LoadBalancer of type `slb.s1.small`

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-spec: "slb.s1.small"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    run: nginx
  type: LoadBalancer
```

#### 7. Attach an exist LoadBalancer to the service with id `${YOUR_LOADBALANCER_ID}`

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-id: "${YOUR_LOADBALANCER_ID}"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    run: nginx
  type: LoadBalancer
```

\>> **Note：** 

- CloudProvider will only help to attach & detach backend server for by default. You need to specify ```service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners: "true"``` to force overwrite listeners. Attention, this might delete the existing listeners.

#### 8. Attach an exist LoadBalancer to the service with id `${YOUR_LOADBALANCER_ID}` , and force to overwrite its listener.

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-id: "${YOUR_LOADBALANCER_ID}"
    service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners: "true"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    run: nginx
  type: LoadBalancer
```

#### 9. Use label to select certain backend for the LoadBalancer.

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-backend-label: "failure-domain.beta.kubernetes.io/zone:ap-southeast-5a"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    run: nginx
  type: LoadBalancer
```

\>> **Note** 

- Separate multiple labels with comma。 "k1:v1,k2:v2"
- And is used in multiple label。

#### 10. Config SessionSticky for TCP LoadBalancer.

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-persistence-timeout: "1800"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    run: nginx
  type: LoadBalancer
```

\>> **Note** 

- Only worked with TCP listeners。
- SessionStichy is applied to all the TCP listeners by default.

#### 11. Config SessionSticky for HTTP & HTTPS LoadBalancer（insert cookie）

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session: "on"
    service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type: "insert"
    service.beta.kubernetes.io/alicloud-loadbalancer-cookie-timeout: "1800"
    service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port: "http:80"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

\>> **Note** 

- Only HTTP & HTTPS。
- SessionSticky type is `insert` Cookie.
- SessionStichy is applied to all the HTTP&HTTPS listeners by default

#### 12. Config SessionSticky for HTTP & HTTPS LoadBalancer（server cookie）

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session: "on"
    service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type: "server"
    service.beta.kubernetes.io/alicloud-loadbalancer-cookie: "${YOUR_COOKIE}"
    service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port: "http:80"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

\>> **Note** 

- Only HTTP & HTTPS。
- SessionSticky type is `server` Cookie.
- SessionStichy is applied to all the HTTP&HTTPS listeners by default

#### 13. Create LoadBalancer with specified master zoneid and slave zoneid

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-master-zoneid: "ap-southeast-5a"
    service.beta.kubernetes.io/alicloud-loadbalancer-slave-zoneid: "ap-southeast-5a"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

\>> **Note** 

- master/slave zone is not supported in every zone，ap-southeast-5 for example does not support master/slave zone.
- modify master/slave available zone is not supported once LoadBalancer has been created.

#### 13. Create Local traffic LoadBalancer

```
apiVersion: v1
kind: Service
metadata:
  name: nginx
  namespace: default
spec:
  externalTrafficPolicy: Local
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

