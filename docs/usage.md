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

**Private Zone** In latest version of cloud controller manager, you can use annotation to bind SLB ip to private zone record. Here is a example:
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: foo
  name: foo-service
  namespace: default
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-private-zone-name: "service.ali"
    service.beta.kubernetes.io/alibaba-cloud-private-zone-record-name: "foo"
spec:
  ports:
    - port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: foo
  type: LoadBalancer
```

To make private zone work, you need config RAM role of K8s master node. Find the role in [RAM control panel](https://ram.console.aliyun.com/roles) and add private zone access permissions to it.

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

```yaml
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

```yaml
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

```yaml
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

```yaml
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

```yaml
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

```yaml
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

```yaml
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

```yaml
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

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-backend-label: "failure-domain.beta.kubernetes.io/region=cn-beijing"
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

- Separate multiple labels with comma。 "k1=v1,k2=v2"
- And is used in multiple label。

#### 10. Config SessionSticky for TCP LoadBalancer.

```yaml
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

```yaml
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

```yaml
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

```yaml
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

#### 14. Create Local traffic LoadBalancer

```yaml
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
#### 15. Create VPC network LoadBalancer 

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-address-type: "intranet"
    service.beta.kubernetes.io/alicloud-loadbalancer-network-type: "vpc"
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
#### 16. Create LoadBalancer charged by bandwidth  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth: "45"
    service.beta.kubernetes.io/alicloud-loadbalancer-charge-type: "paybybandwidth"
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
#### 17. Create specific region LoadBalacer  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-region: "cn-beijing"
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
#### 18. Create LoadBalancer with health check 

- TCP Type

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-tag: "on"
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-type: "tcp"
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout: "8"
    service.beta.kubernetes.io/alicloud-loadbalancer-healthy-threshold: "4"
    service.beta.kubernetes.io/alicloud-loadbalancer-unhealthy-threshold: "4"
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval: "3"
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
- HTTP Type

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-flag: "on"
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-type: "http"
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-uri: "/test/index.html"
    service.beta.kubernetes.io/alicloud-loadbalancer-healthy-threshold: "4"
    service.beta.kubernetes.io/alicloud-loadbalancer-unhealthy-threshold: "4"
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout: "10"
    service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval: "3"
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

- All of the above parameters are mandatory for http type.

#### 19. Create LoadBalancer with wlc scheduler  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-scheduler: "wlc"
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
#### 20. Create LoadBalancer with acl  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-status: "on"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-id: "acl-2zeckgpq7xxx1hbdsxxxx"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-type: "white"
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
#### 21. Create LoadBalancer with specific vswitchid  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
   service.beta.kubernetes.io/alicloud-loadbalancer-address-type: "intranet"
   service.beta.kubernetes.io/alicloud-loadbalancer-vswitch-id: "vsw-2zewxxxxgr3xhfl2xxxxx"
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
#### 22. Create LoadBalacer with forward port  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port: "http:80"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-forward-port: "81"
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
#### 23. Create LoadBalacer with additional resource tags  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags: "Key1:Value1,Key2:Value2" 
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
#### 24. Create ipv6 LoadBalancer  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-ip-version: "ipv6"
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

#### 25. Create LoadBalancer with PrivateZone  

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: foo
  name: foo-service
  namespace: default
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-private-zone-name: "service.ali"
    service.beta.kubernetes.io/alibaba-cloud-private-zone-record-name: "foo"
    service.beta.kubernetes.io/alibaba-cloud-private-zone-record-ttl: "10"
spec:
  ports:
    - port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: foo
  type: LoadBalancer
```


ps: Annotation list

|Annotation |Description |Default value |
|----|----|---|
|service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port|	Use a commas (,) to separate two values, for example, https:443,http:80.	|None|
|service.beta.kubernetes.io/alicloud-loadbalancer-address-type|	Valid values: internet or intranet.	|internet|
|service.beta.kubernetes.io/alicloud-loadbalancer-slb-network-type|The network type of the SLB instance can be classic or vpc.|classic|
|service.beta.kubernetes.io/alicloud-loadbalancer-charge-type|	Valid values: paybytraffic or paybybandwidth.	|paybytraffic|
|service.beta.kubernetes.io/alicloud-loadbalancer-id|	ID of the SLB instance.<br/>Specify an existing SLB instance you own through *service.beta.kubernetes.io/alicloud-loadbalancer-id* and the existing listeners are overridden. <br/>Note that the SLB instance is not deleted when you delete the service.	|None|
|service.beta.kubernetes.io/alicloud-loadbalancer-backend-label|	Use labels to specify the Worker nodes to be mounted to the backend of the SLB instance.	|None|
|service.beta.kubernetes.io/alicloud-loadbalancer-spec|	Specification of the SLB instance. For more information, see [CreateLoadBalancer](https://www.alibabacloud.com/help/doc-detail/27577.htm?#SLB-api-CreateLoadBalancer)	|None|
|service.beta.kubernetes.io/alicloud-loadbalancer-persistence-timeout| Session timeout period. It applies only to TCP listeners and the value range is 0 to 3600 (seconds). The default value is 0, indicating that the session remains closed. For more information, see [CreateLoadBalancerTCPListener](https://www.alibabacloud.com/help/doc-detail/27594.htm?#slb-api-CreateLoadBalancerTCPListener). |0|
|service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session| Whether to enable session persistence. <br/>Valid values: on or off. <br/>**Note** It applies only to HTTP and HTTPS listeners.<br/> For more information, see [CreateLoadBalancerHTTPListener](https://www.alibabacloud.com/help/doc-detail/27592.htm?#slb-api-CreateLoadBalancerHTTPListener) and [CreateLoadBalancerHTTPSListener](https://www.alibabacloud.com/help/doc-detail/27593.htm?#slb-api-CreateLoadBalancerHTTPSListener). |off|
|service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type| Method used to handle the cookie. <br/>Valid values: <br> - insert: Insert the cookie. <br> - server: Rewrite the cookie.<br> Note It applies only to HTTP and HTTPS listeners.When the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session* is set to on, this parameter is mandatory. |None|
|service.beta.kubernetes.io/alicloud-loadbalancer-cookie-timeout| Timeout period of the cookie.<br> Value range: 1–8640 (seconds).<br/>**Note** When the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session* is set to on and the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type* is set to insert, this parameter is mandatory. |None|
|service.beta.kubernetes.io/alicloud-loadbalancer-cookie| Cookie configured on the server. <br>The cookie must be 1 to 200 characters in length and can only contain ASCII English letters and numeric characters. It cannot contain commas, semicolons, or spaces, or begin with $.<br>**Note**  When the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session* is set to on and the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type* is set to server, this parameter is mandatory. |None|
|service.beta.kubernetes.io/alicloud-loadbalancer-master-zoneid|	Availability zone ID of the primary backend server.	|None|
|service.beta.kubernetes.io/alicloud-loadbalancer-slave-zoneid|	Availability zone ID of the secondary backend server.	|None|
|externalTrafficPolicy| Nodes that can be used as backend servers. <br/>Valid values:<br/>**Cluster**: Use all backend nodes as backend servers.<br>**Local**: Use the nodes where pods are located as backend servers. |Cluster|
|service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners|	Whether to forcibly override the listeners when you specify an existing SLB instance.	|false: Do not override.|
|service.beta.kubernetes.io/alicloud-loadbalancer-region|	Rgion where the SLB instance is located.	|None|
|service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth|	Bandwidth of the SLB instance.	|50|
|service.beta.kubernetes.io/alicloud-loadbalancer-cert-id|	ID of a certificate on Alibaba Cloud. You must have uploaded a certificate first.	|None|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-flag|	Valid values: on or off.	|The default value is off. No need to modify this parameter for TCP, because health check is enabled for TCP by default and this parameter cannot be set.|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-type|	Health check type. <br/>Valid values: tcp or http. |tcp|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-uri| URI used for health check. <br>**Note** If the health check type is TCP, you do not need to set this parameter. |None|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-port|	Port used for health check.<br> Valid values:<br/>  **-520**: The backend port configured for the listener is used by default.<br>  **1-65535**: The port opened on the backend server for health check is used. |None|
|service.beta.kubernetes.io/alicloud-loadbalancer-healthy-threshold| The number of consecutive health check successes before the backend server is deemed as healthy (from failure to success). <br>Value range: 2–10. <br>For more information, see CreateLoadBalancerTCPListener. |3|
|service.beta.kubernetes.io/alicloud-loadbalancer-unhealthy-threshold|	The number of consecutive health check fails before the backend server is deemed as unhealthy (from success to failure). <br>Value range: 2–10. |3|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval| Time interval between two consecutive health checks.<br> Value range: 1–50 (seconds). |2|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout| Amount of time waiting for the response from TCP type health check. If the backend ECS instance does not send a valid response within a specified period of time, the health check fails. <br>value range: 1–300 (seconds).<br>**Note** If the value of the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout* is less than that of the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval*, the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout* is invalid and the timeout period equals the value of *service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval*. |5|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout| Amount of time waiting for the response from HTTP type health check. If the backend ECS instance does not send a valid response within a specified period of time, the health check fails.<br>Value range: 1–300 (seconds).<br>**Note** If the value of the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout*is less than that of the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval*, the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout* is invalid, and the timeout period equals the value of the parameter *service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval*. |5|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-domain|The domain used for health checks. <br>Valid values:<br>**$_ip**: Private network IP of the backend server. When IP is specified or the parameter is not specified, load balancer uses the private network IP of each backend server as the domain used for health check.<br>**domain**: The length of domain is between 1-80 characters and can only contain letters, numbers, periods (.) and hyphens (-).|None|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-httpcode|Normal HTTP status codes for the health check.<br> Multiple status codes are separated by commas (,).<br>Valid values: http_2xx, http_3xx, http_4xx or http_5xx.|http_2xx|
|service.beta.kubernetes.io/alibaba-cloud-private-zone-name|Name of PrivateZone, cloud-controller will not to create new PrivateZone.|None|
|service.beta.kubernetes.io/alibaba-cloud-private-zone-id|Id of PrivateZone, when it was configured with name together, cloud provider will find PrivateZone by id first.|None|
|service.beta.kubernetes.io/alibaba-cloud-private-zone-record-name|Name of PrivateZone record.|None|
|service.beta.kubernetes.io/alibaba-cloud-private-zone-record-ttl|TTL of PrivateZone record, 60s as default value.|60|
|service.beta.kubernetes.io/alicloud-loadbalancer-scheduler|The scheduling algorithm.<br> Valid values: wrr or wlc or rr. <br>**wrr** (default): The higher the weight value of the backend server, the higher the number of polls (probability). <br>**wlc**: In addition to polling based on the weight value set by each back-end server, the actual load of the back-end server (ie, the number of connections) is also considered. When the weight values are the same, the smaller the number of current connections, the higher the number of times (probability) that the backend server is polled.<br>**rr**: The external requests are sequentially distributed to the backend server in order of access.|wrr|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-status|Whether to enable access control. <br>Valid values: on or off.|off|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-id|Access control ID.<br/>**Note** If the value of AclStatus is "on", this parameter must be set.|None|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-type|The types of access control.<br>Valid values: white or black.<br>**white**：Only requests from IP addresses or address segments set in the selected access control policy group are forwarded. The whitelist is suitable for scenarios where the application only allows specific IP access.Note Once the whitelist is set, only the IPs in the whitelist can access the load balancing listener. If whitelist access is turned on, but no IP is added to the access policy group, the load balancing listener forwards all requests.<br/>**black**： All requests from the IP address or address segment set in the selected access control policy group are not forwarded. The blacklist is suitable for scenarios where the application only rejects certain IPs access.Note If blacklist access is turned on, but no IP is added to the access policy group, the load balancing listener forwards all requests.<br>If the value of AclStatus is "on", this parameter must be set.|None|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-vswitch-id|VSwitch ID of the load balancer.<br>Note When setting VSwitch ID, the address-type parameter need to be "intranet".|None|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-forward-port|HTTP to HTTPS listening forwarding port.|80|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags|A list of tags to add.<br>e.g. "k1:v1,k2:v2"|None|
|service.beta.kubernetes.io/alibaba-cloud-loadbalancer-ip-version|IP version of the load balancer.<br/>Valid values：ipv4 or ipv6.|ipv4|


