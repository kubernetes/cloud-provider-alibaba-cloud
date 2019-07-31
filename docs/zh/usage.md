# 通过负载均衡（Server Load Balancer）访问服务

您可以使用阿里云负载均衡来访问服务。

## 背景信息

如果您的集群的cloud-controller-manager版本大于等于v1.9.3，对于指定已有SLB的时候，系统默认不再为该SLB处理监听，用户需要手动配置该SLB的监听规则。

执行以下命令，可查看cloud-controller-manager的版本。

```
root@master # kubectl get po -n kube-system -o yaml|grep image:|grep cloud-con|uniq

image: registry-vpc.cn-hangzhou.aliyuncs.com/acs/cloud-controller-manager-amd64:v1.9.3
```

## 通过命令行操作

1. 通过命令行工具创建一个 Nginx 应用。

   ```bash
   root@master # kubectl run nginx -image=registry.aliyuncs.com/acs/netdia:latest
   root@master # kubectl get po
   NAME                                   READY     STATUS    RESTARTS   AGE
   nginx-2721357637-dvwq3                 1/1       Running   1          6s
   ```

2. 为 Nginx 应用创建阿里云负载均衡服务，指定 `type=LoadBalancer` 来向外网用户暴露 Nginx 服务。

   ```bash
   root@master # kubectl expose deployment nginx --port=80 --target-port=80 --type=LoadBalancer
   root@master # kubectl get svc
   NAME                  CLUSTER-IP      EXTERNAL-IP      PORT(S)                        AGE
   nginx                 172.19.10.209   101.37.192.20   80:31891/TCP                   4s
   ```

3. 在浏览器中访问 `http://101.37.192.20`，来访问您的 Nginx 服务。

## 负载均衡与 Private Zone 的绑定

在新版的cloud-controller-manager中支持了在配置文件中自动将SLB与Private Zone的解析记录绑定。

该功能能够为自动创建的SLB能够在内网中与某个域名绑定起来，方便自动创建的SLB的管理。具体的Private Zone服务请参考[什么是Private Zone](https://help.aliyun.com/document_detail/64611.html)。

要使用本功能，还需要授予k8s集群的Master节点相应的RAM授权，让Master节点能够访问Private Zone服务API。

需要在[RAM 角色管理控制台](https://ram.console.aliyun.com/roles)中找到Master节点的RAM角色，并给与PrivateZone的访问权。

## 更多信息

阿里云负载均衡还支持丰富的配置参数，包含健康检查、收费类型、负载均衡类型等参数。详细信息参见[负载均衡配置参数表]。

## 注释

阿里云可以通过注释`annotations`的形式支持丰富的负载均衡功能。

- 保存以下yaml为svc.1.yaml ， 然后使用 kubectl apply -f svc.1.yaml的方式来创建service.

**1. 创建一个公网类型的负载均衡**

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

**2. 创建一个私网类型的负载均衡**

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

**3. 创建HTTP类型的负载均衡**

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

**4. 创建HTTPS类型的负载均衡**

- 需要先在阿里云控制台上创建一个证书并记录 cert-id，然后使用如下 annotation 创建一个 HTTPS 类型的 SLB。

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

**5. 限制负载均衡的带宽**

- 只限制负载均衡实例下的总带宽，所有监听共享实例的总带宽，参见[共享实例带宽](https://help.aliyun.com/document_detail/85930.html)。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-charge-type: "paybybandwidth"
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

**6. 指定负载均衡规格**

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

**7. 使用已有的负载均衡**

- 默认情况下，使用已有的负载均衡实例，不会覆盖监听，如要强制覆盖已有监听，请配置service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners为true。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-id: "your_loadbalancer_id"
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

**8. 使用已有的负载均衡，并强制覆盖已有监听**

- 强制覆盖已有监听，会删除已有负载均衡实例上的已有监听。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-id: "your_loadbalancer_id"
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
  type: LoadBalancere: LoadBalancer
```

**9. 使用指定label的worker节点作为后端服务器**

- 多个label以逗号分隔。例如："k1=v1,k2=v2"。多个label之间是`and`的关系。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-backend-label: "failure-domain.beta.kubernetes.io/zone=ap-southeast-5a"
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

**10. 为TCP类型的负载均衡配置会话保持保持时间**

- 该参数service.beta.kubernetes.io/alicloud-loadbalancer-persistence-tim仅对TCP协议的监听生效。
- 如果负载均衡实例配置了多个TCP协议的监听端口，则默认将该配置应用到所有TCP协议的监听端口。

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

**11. 为HTTP&HTTPS协议的负载均衡配置会话保持（insert cookie）**

- 仅支持HTTP及HTTPS协议的负载均衡实例。
- 如果配置了多个HTTP或者HTTPS的监听端口，该会话保持默认应用到所有HTTP和HTTPS监听端口。

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

**12. 为HTTP&HTTPS协议的负载均衡配置会话保持（server cookie）**

- 仅支持HTTP及HTTPS协议的负载均衡实例。
- 如果配置了多个HTTP或者HTTPS的监听端口，该会话保持默认应用到所有HTTP和HTTPS监听端口。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session: "on"
    service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type: "server"
    service.beta.kubernetes.io/alicloud-loadbalancer-cooyour_cookie: "${your_cookie}"
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

**13. 创建负载均衡时，指定主备可用区**

- 某些region的负载均衡不支持主备可用区，如ap-southeast-5。
- 一旦创建，主备可用区不支持修改。

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

**14. 使用Pod所在的节点作为后端服务器**

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

**15. 创建VPC类型的负载均衡**

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

**16. 创建按流量付费的负载均衡** 

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

**17. 创建指定地域的负载均衡**  

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

**18. 创建带健康检查的负载均衡** 

- 健康检查为TCP类型

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

- 健康检查为HTTP类型

  — 如果健康检查为HTTP类型，则下述所有参数必填。

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

**19. 创建调查算法为加权最小连接数的负载均衡** 

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

**20. 创建带有访问控制的负载均衡**  

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

**21. 为负载均衡指定虚拟交换机**  

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

**22. 为负载均衡指定转发端口**  

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

**23. 为负载均衡添加额外标签**  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags: "Key1=Value1,Key2=Value2" 
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

**24. 创建ipv6类型的负载均衡** 

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

**说明：** 注解的内容是区分大小写的。

| 注释                                                         | 描述                                                         | 默认值                                                       |
| :----------------------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port | 多个值之间由逗号分隔，比如：https:443,http:80                | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-address-type | 取值可以是internet或者intranet                               | internet                                                     |
| service.beta.kubernetes.io/alicloud-loadbalancer-slb-network-type | 负载均衡的网络类型，取值：classic或vpc。<br>取值为vpc时，需设置service.beta.kubernetes.io/alicloud-loadbalancer-address-type为intranet。 | classic                                                      |
| service.beta.kubernetes.io/alicloud-loadbalancer-charge-type | 取值可以是paybytraffic或者paybybandwidth                     | paybytraffic                                                 |
| service.beta.kubernetes.io/alicloud-loadbalancer-id          | 负载均衡实例的 ID。通过 service.beta.kubernetes.io/alicloud-loadbalancer-id指定您已有的SLB，已有监听会被覆盖， 删除 service 时该 SLB 不会被删除。 | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-backend-label | 通过 label 指定 SLB 后端挂载哪些worker节点。                 | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-spec        | 负载均衡实例的规格。可参考：[文档](https://help.aliyun.com/document_detail/27577.html?spm=a2c4g.11186623.2.22.6be5609abM821g#SLB-api-CreateLoadBalancer) | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-persistence-timeout | 会话保持时间。<br>仅针对TCP协议的监听，取值：0-3600（秒），默认情况下，取值为0，会话保持关闭。<br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener) | 0                                                            |
| service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session | 是否开启会话保持。取值：on                                   | off<br>**说明：** 仅对HTTP和HTTPS协议的监听生效。<br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)<br> |
| service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type | cookie的处理方式。<br>取值：<br>**insert**：植入Cookie。<br>**server**：重写Cookie。<br>**说明：**  <br>- 仅对HTTP和HTTPS协议的监听生效。 <br>- 当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session取值为on时，该参数必选。<br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)<br> | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-cookie-timeout | Cookie超时时间。取值：1-86400（秒）<br>**说明：** 当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session为on且service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type为insert时，该参数必选。<br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)<br> | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-cookie      | 服务器上配置的Cookie。长度为1-200个字符，只能包含ASCII英文字母和数字字符，不能包含逗号、分号或空格，也不能以$开头。<br>**说明：** <br>当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session为on且service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type为server时，该参数必选。<br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)<br> | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-master-zoneid | 主后端服务器的可用区ID。                                     | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-slave-zoneid | 备后端服务器的可用区ID。                                     | 无                                                           |
| externalTrafficPolicy                                        | 哪些节点可以作为后端服务器，取值：<br>**Cluster**：使用所有后端节点作为后端服务器。<br> **Local**：使用Pod所在节点作为后端服务器。 | Cluster                                                      |
| service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners | 绑定已有负载均衡时，是否强制覆盖该SLB的监听。                | false：不覆盖                                                |
| service.beta.kubernetes.io/alicloud-loadbalancer-region      | 负载均衡所在的地域                                           | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth   | 负载均衡的带宽                                               | 50                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-cert-id     | 阿里云上的证书 ID。您需要先上传证书                          | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-flag | 取值是on                                                     | off                                                          |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-type | 健康检查类型，取值：tcp或http。可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br> | tcp                                                          |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-uri | 用于健康检查的URI。**说明：** 当健康检查类型为TCP模式时，无需配置该参数。<br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br> | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-port | 健康检查使用的端口。<br>取值：<br>**-520**：默认使用监听配置的后端端口。<br> **1-65535**：健康检查的后端服务器的端口。<br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br> | 无                                                           |
| service.beta.kubernetes.io/alicloud-loadbalancer-healthy-threshold | 健康检查连续成功多少次后，将后端服务器的健康检查状态由fail判定为success。<br>取值：2-10<br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br> | 3                                                            |
| service.beta.kubernetes.io/alicloud-loadbalancer-unhealthy-threshold | 健康检查连续失败多少次后，将后端服务器的健康检查状态由success判定为fail。<br>取值：2-10<br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br> | 3                                                            |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval | 健康检查的时间间隔。<br> 取值：1-50（秒）<br> 可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br> | 2                                                            |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout | 接收来自运行状况检查的响应需要等待的时间,适用于TCP模式。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。<br> 取值：1-300（秒）<br> **说明：** 如果service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout的值小于service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值，则service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout无效，超时时间为service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值。<br> 可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br> | 5                                                            |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout | 接收来自运行状况检查的响应需要等待的时间，适用于HTTP模式。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。取值：1-300（秒）<br>**说明：** 如果 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout的值小于service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值，则 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout无效，超时时间为 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值。<br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)<br> | 5                                                            |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-domain | 用于健康检查的域名。<br>**$_ip**： 后端服务器的私网IP。当指定了IP或该参数未指定时，负载均衡会使用各后端服务器的私网IP当做健康检查使用的域名。<br>**domain**：域名长度为1-80，只能包含字母、数字、点号（.）和连字符（-）。 | 无                                                           |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-httpcode | 健康检查正常的HTTP状态码，多个状态码用逗号（,）分割。<br>取值：http_2xx（默认值）或http_3xx或http_4xx或http_5xx。 | http_2xx                                                     |
| service.beta.kubernetes.io/alicloud-loadbalancer-scheduler   | 调度算法。取值 wrr或wlc或rr。<br>**wrr**（默认值）：权重值越高的后端服务器，被轮询到的次数（概率）也越高。<br>**wlc**：除了根据每台后端服务器设定的权重值来进行轮询，同时还考虑后端服务器的实际负载（即连接数）。当权重值相同时，当前连接数越小的后端服务器被轮询到的次数（概率）也越高。<br>**rr**：按照访问顺序依次将外部请求依序分发到后端服务器。 | wrr                                                          |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-status | 是否开启访问控制功能。取值：on或off                          | off                                                          |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-id | 监听绑定的访问策略组ID。当AclStatus参数的值为on时，该参数必选。 | 无                                                           |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-type | 访问控制类型。<br>取值：white或black。<br>**white**： 仅转发来自所选访问控制策略组中设置的IP地址或地址段的请求，白名单适用于应用只允许特定IP访问的场景。设置白名单存在一定业务风险。一旦设名单，就只有白名单中的IP可以访问负载均衡监听。如果开启了白名单访问，但访问策略组中没有添加任何IP，则负载均衡监听会转发全部请求。<br>**black**： 来自所选访问控制策略组中设置的IP地址或地址段的所有请求都不会转发，黑名单适用于应用只限制某些特定IP访问的场景。如果开启了黑名单访问，但访问策略组中没有添加任何IP，则负载均衡监听会转发全部请求。当AclStatus参数的值为on时，该参数必选。 | 无                                                           |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-vswitch-id | 负载均衡实例所属的VSwitch ID。设置改参数时需同时设置addresstype为intranet。 | 无                                                           |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-forward-port | HTTP至HTTPS的监听转发端口。                                  | 80                                                           |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags | 需要添加的Tag列表。如："k1=v1,k2=v2"                         | 无                                                           |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-ip-version | 负载均衡实例的IP版本，取值：ipv4或ipv6。                     | ipv4                                                         |

