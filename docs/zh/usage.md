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


1.  通过命令行工具创建一个 Nginx 应用。

    ```
    root@master # kubectl run nginx --image=registry.aliyuncs.com/acs/netdia:latest
    root@master # kubectl get po 
    NAME                                   READY     STATUS    RESTARTS   AGE
    nginx-2721357637-dvwq3                 1/1       Running   1          6s
    ```

2.  为 Nginx 应用创建阿里云负载均衡服务，指定 `type=LoadBalancer` 来向外网用户暴露 Nginx 服务。

    ```
    root@master # kubectl expose deployment nginx --port=80 --target-port=80 --type=LoadBalancer
    root@master # kubectl get svc
    NAME                  CLUSTER-IP      EXTERNAL-IP      PORT(S)                        AGE
    nginx                 172.19.10.209   101.37.192.20   80:31891/TCP                   4s
    ```

3.  在浏览器中访问 `http://101.37.192.20`，来访问您的 Nginx 服务。

## 更多信息 

阿里云负载均衡还支持丰富的配置参数，包含健康检查、收费类型、负载均衡类型等参数。详细信息参见[负载均衡配置参数表]。

## 注释 

阿里云可以通过注释`annotations`的形式支持丰富的负载均衡功能。

**使用已有的内网 SLB**

需要指定两个annotation。注意修改成您自己的 Loadbalancer-id。

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-address-type: "intranet"
    service.beta.kubernetes.io/alicloud-loadbalancer-id: "your-loadbalancer-id"
  labels:
    run: nginx
  name: nginx
  namespace: default
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  sessionAffinity: None
  type: LoadBalancer
```

**创建HTTP类型的负载均衡**

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

**创建 HTTPS 类型的负载均衡**

需要先在阿里云控制台上创建一个证书并记录 cert-id，然后使用如下 annotation 创建一个 HTTPS 类型的 SLB。

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-cert-id: "your-cert-id"
    service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port: "https:443"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    run: nginx
  sessionAffinity: None
  type: LoadBalancer
```

**限制负载均衡的带宽**

只限制负载均衡实例下的总带宽，所有监听共享实例的总带宽，参见[共享实例带宽](../../../../intl.zh-CN/用户指南/监听/共享实例带宽.md#)。

```
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

**指定负载均衡规格**

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

**使用已有的负载均衡**

默认情况下，使用已有的负载均衡实例，不会覆盖监听，如要强制覆盖已有监听，请配置service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners为true。

```
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
  type: LoadBalancer: LoadBalancer
```

****使用已有的负载均衡，并强制覆盖已有监听****

强制覆盖已有监听，会删除已有负载均衡实例上的已有监听。

```
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

****使用指定label的worker节点作为后端服务器****

多个label以逗号分隔。例如："k1:v1,k2:v2"

多个label之间是`and`的关系。

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

**为TCP类型的负载均衡配置会话保持保持时间**

该参数service.beta.kubernetes.io/alicloud-loadbalancer-persistence-tim仅对TCP协议的监听生效。

如果负载均衡实例配置了多个TCP协议的监听端口，则默认将该配置应用到所有TCP协议的监听端口。

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

**为HTTP&HTTPS协议的负载均衡配置会话保持（insert cookie）**

仅支持HTTP及HTTPS协议的负载均衡实例。

如果配置了多个HTTP或者HTTPS的监听端口，该会话保持默认应用到所有HTTP和HTTPS监听端口。

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

**为HTTP&HTTPS协议的负载均衡配置会话保持（server cookie）**

仅支持HTTP及HTTPS协议的负载均衡实例。

如果配置了多个HTTP或者HTTPS的监听端口，该会话保持默认应用到所有HTTP和HTTPS监听端口。

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session: "on"
    service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type: "server"
    service.beta.kubernetes.io/alicloud-loadbalancer-cooyour_cookie: "your_cookie"
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

**创建负载均衡时，指定主备可用区**

某些region的负载均衡不支持主备可用区，如ap-southeast-5。

一旦创建，主备可用区不支持修改。

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

**使用Pod所在的节点作为后端服务器**

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

**说明：** 注释的内容是区分大小写的。

|注释|描述|默认值|
|--|--|---|
|service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port|多个值之间由逗号分隔，比如：https:443,http:80|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-address-type|取值可以是internet或者intranet|internet|
|service.beta.kubernetes.io/alicloud-loadbalancer-slb-network-type|负载均衡的网络类型，取值可以是classic或者vpc|classic|
|service.beta.kubernetes.io/alicloud-loadbalancer-charge-type|取值可以是paybytraffic或者paybybandwidth|paybytraffic|
|service.beta.kubernetes.io/alicloud-loadbalancer-id|负载均衡实例的 ID。通过 service.beta.kubernetes.io/alicloud-loadbalancer-id指定您已有的SLB，已有监听会被覆盖， 删除 service 时该 SLB 不会被删除。|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-backend-label|通过 label 指定 SLB 后端挂载哪些worker节点。|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-spec|负载均衡实例的规格。可参考：[文档](https://help.aliyun.com/document_detail/27577.html?spm=a2c4g.11186623.2.22.6be5609abM821g#SLB-api-CreateLoadBalancer)|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-persistence-timeout|会话保持时间。仅针对TCP协议的监听，取值：0-3600（秒）</br></br>默认情况下，取值为0，会话保持关闭。</br></br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)</br></br>|0|
|service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session|是否开启会话保持。取值：on | off**说明：** 仅对HTTP和HTTPS协议的监听生效。</br></br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)</br></br>|off|
|service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type|cookie的处理方式。取值：-   insert：植入Cookie。</br>-   server：重写Cookie。</br></br>**说明：** </br></br>-   仅对HTTP和HTTPS协议的监听生效。</br>-   当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session取值为on时，该参数必选。</br></br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)</br></br>|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-cookie-timeout|Cookie超时时间。取值：1-86400（秒）**说明：** 当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session为on且service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type为insert时，该参数必选。</br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)</br></br>|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-cookie|服务器上配置的Cookie。长度为1-200个字符，只能包含ASCII英文字母和数字字符，不能包含逗号、分号或空格，也不能以$开头。</br></br>**说明：** </br></br>当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session为on且service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type为server时，该参数必选。</br></br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)</br></br>|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-master-zoneid|主后端服务器的可用区ID。|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-slave-zoneid|备后端服务器的可用区ID。|无|
|externalTrafficPolicy|哪些节点可以作为后端服务器，取值：-   Cluster：使用所有后端节点作为后端服务器。</br></br>-   Local：使用Pod所在节点作为后端服务器。</br></br></br>|Cluster|
|service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners|绑定已有负载均衡时，是否强制覆盖该SLB的监听。|false：不覆盖|
|service.beta.kubernetes.io/alicloud-loadbalancer-region|负载均衡所在的地域|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth|负载均衡的带宽|50|
|service.beta.kubernetes.io/alicloud-loadbalancer-cert-id|阿里云上的证书 ID。您需要先上传证书|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-flag|取值是on | off|默认为off。TCP 不需要改参数。因为 TCP 默认打开健康检查，用户不可设置。|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-type|健康检查类型，取值：tcp | http。可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)</br></br>|tcp|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-uri|用于健康检查的URI。**说明：** 当健康检查类型为TCP模式时，无需配置该参数。</br></br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)</br></br>|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-port|健康检查使用的端口。取值：-   -520：默认使用监听配置的后端端口。</br>-   1-65535：健康检查的后端服务器的端口。</br></br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)</br></br>|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-healthy-threshold| 可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)</br></br> |无|
|service.beta.kubernetes.io/alicloud-loadbalancer-unhealthy-threshold|健康检查连续成功多少次后，将后端服务器的健康检查状态由fail判定为success。取值：2-10</br></br>可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)</br></br>|无|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval| 健康检查的时间间隔。</br></br> 取值：1-50（秒）</br></br> 可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)</br></br> |无|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout| 接收来自运行状况检查的响应需要等待的时间。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。</br></br> 取值：1-300（秒）</br></br> **说明：** 如果service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout的值小于service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值，则service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout无效，超时时间为service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值。</br></br> 可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)</br></br> |无|
|service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout|接收来自运行状况检查的响应需要等待的时间。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。取值：1-300（秒）</br></br>**说明：** 如果 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout的值小于service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值，则 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout无效，超时时间为 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值。</br></br>可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)</br></br>|无|
