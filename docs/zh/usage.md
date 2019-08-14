# Cloud Provider 帮助文档（中文版）

# 通过负载均衡（Server Load Balancer）访问服务

您可以使用阿里云负载均衡来访问服务。

## 背景信息

如果您的集群的cloud-controller-manager版本大于等于v1.9.3，对于指定已有SLB的时候，系统默认不再为该SLB处理监听，用户需要手动配置该SLB的监听规则。

执行以下命令，可查看cloud-controller-manager的版本。

```
root@master # kubectl get po -n kube-system -o yaml|grep image:|grep cloud-con|uniq

image: registry-vpc.cn-hangzhou.aliyuncs.com/acs/cloud-controller-manager-amd64:v1.9.3
```

## 注意事项

- Cloud Controller Manager（简称CCM）会为`Type=LoadBalancer`类型的Service创建或配置阿里云负载均衡（SLB），包含**SLB**、**监听**、**虚拟服务器组**等资源。
- 对于非LoadBalancer类型的service则不会为其配置负载均衡，这包含如下场景：当用户将`Type=LoadBalancer`的service变更为`Type!=LoadBalancer`时，CCM也会删除其原先为该Service创建的SLB（用户通过`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id annotation`指定的已有SLB除外）。
- 自动刷新配置：CCM使用声明式API，会在一定条件下自动根据service的配置刷新阿里云负载均衡配置，所有用户自行在SLB控制台上修改的配置均存在被覆盖的风险（使用已有SLB同时不覆盖监听的场景除外），因此不能在SLB控制台手动修改Kubernetes创建并维护的SLB的任何配置，否则有配置丢失的风险。
- 同时支持为serivce指定一个已有的负载均衡，或者让CCM自行创建新的负载均衡。但两种方式在SLB的管理方面存在一些差异：指定已有SLB
  - 需要为Service设置`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id` annotation。
  - SLB配置：此时CCM会使用该SLB做为Service的SLB，并根据其他annotation配置SLB，并且自动的为SLB创建多个虚拟服务器组（当集群节点变化的时候，也会同步更新虚拟服务器组里面的节点）。
  - 监听配置：是否配置监听取决于`service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners: `是否设置为true。 如果设置为false，那么CCM不会为SLB管理任何监听配置。如果设置为true，那么CCM会尝试为SLB更新监听，此时CCM会根据监听名称判断SLB上的监听是否为k8s维护的监听（名字的格式为k8s/Port/ServiceName/Namespace/ClusterID），若Service声明的监听与用户自己管理的监听端口冲突，那么CCM会报错。
  - SLB的删除： 当Service删除的时候CCM不会删除用户通过id指定的已有SLB。
- CCM管理的SLB<br />
  - CCM会根据service的配置自动的创建配置**SLB**、**监听**、**虚拟服务器组**等资源，所有资源归CCM管理，因此用户不得手动在SLB控制台更改以上资源的配置，否则CCM在下次Reconcile的时候将配置刷回service所声明的配置，造成非用户预期的结果。
  - SLB的删除：当Service删除的时候CCM会删除该SLB。
- 后端服务器更新
  - CCM会自动的为该Service对应的SLB刷新后端虚拟服务器组。当Service对应的后端Endpoint发生变化的时候或者集群节点变化的时候都会自动的更新SLB的后端Server。
  - `spec.ExternalTraffic = Cluster`模式的Service，CCM默认会将所有节点挂载到SLB的后端（使用BackendLabel标签配置后端的除外）。由于SLB限制了每个ECS上能够attach的SLB的个数（quota），因此这种方式会快速的消耗该quota,当quota耗尽后，会造成Service Reconcile失败。解决的办法，可以使用Local模式的Service。
  - `spec.ExternalTraffic = Local`模式的Service，CCM默认只会讲Service对应的Pod所在的节点加入到SLB后端。这会明显降低quota的消耗速度。同时支持四层源IP保留。
  - 任何情况下CCM不会将Master节点作为SLB的后端。
  - CCM会从SLB后端摘除被kubectl drain/cordon的节点。

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

## 更多信息

阿里云负载均衡还支持丰富的配置参数，包含健康检查、收费类型、负载均衡类型等参数。

## 注释

阿里云可以通过注释`annotations`的形式支持丰富的负载均衡功能。

- **创建一个公网类型的负载均衡**

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

- **创建一个私网负载的均衡**

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

- **创建HTTP类型的负载均衡**

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

- **创建HTTPS类型的负载均衡**

需要先在阿里云控制台上创建一个证书并记录 cert-id，然后使用如下 annotation 创建一个 HTTPS 类型的 SLB。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port: "https:443"
    service.beta.kubernetes.io/alicloud-loadbalancer-cert-id: "${YOUR_CERT_ID}"
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

- **限制负载均衡的带宽**

只限制负载均衡实例下的总带宽，所有监听共享实例的总带宽，参见[共享实例带宽](https://help.aliyun.com/document_detail/85930.html)。

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

- **指定负载均衡规格**

  负载均衡规格可参考[文档](https://help.aliyun.com/document_detail/27577.html?spm=a2c4g.11186623.2.22.6be5609abM821g#SLB-api-CreateLoadBalancer)。

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

- **使用已有的负载均衡**

默认情况下，使用已有的负载均衡实例，不会覆盖监听，如要强制覆盖已有监听，请配置service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners为true。<br />使用已有的负载均衡暂不支持添加额外标签（annotation: service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags）
<br />复用已有的负载均衡默认不覆盖已有监听，出于以下两点原因：
1）如果已有负载均衡的监听上绑定了业务，强制覆盖会引发业务中断
2）由于CCM目前支持的后端配置有限，无法处理一些复杂配置。如果有复杂的后端配置需求，可以通过手动方式自行配置。

如存在以上两种情况不建议强制覆盖监听，如果已有负载均衡的监听端口不在使用，则可以强制覆盖。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-id: "${YOUR_LOADBALACER_ID}"
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

- **使用已有的负载均衡，并强制覆盖已有监听**

强制覆盖已有监听，如果监听端口冲突，则会删除已有监听。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-id: "${YOUR_LOADBALACER_ID}"
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

- **使用指定label的worker节点作为后端服务器**

多个label以逗号分隔。例如："k1=v1,k2=v2"。多个label之间是`and`的关系。

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

- **为TCP类型的负载均衡配置会话保持保持时间**

参数service.beta.kubernetes.io/alicloud-loadbalancer-persistence-time仅对TCP协议的监听生效。<br />如果负载均衡实例配置了多个TCP协议的监听端口，则默认将该配置应用到所有TCP协议的监听端口。

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

- **为HTTP&HTTPS协议的负载均衡配置会话保持（insert cookie）**

仅支持HTTP及HTTPS协议的负载均衡实例。<br />如果配置了多个HTTP或者HTTPS的监听端口，该会话保持默认应用到所有HTTP和HTTPS监听端口。<br />配置insert cookie，以下四项annotation必选。

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

- **为HTTP&HTTPS协议的负载均衡配置会话保持（server cookie）**

仅支持HTTP及HTTPS协议的负载均衡实例。<br />如果配置了多个HTTP或者HTTPS的监听端口，该会话保持默认应用到所有HTTP和HTTPS监听端口。<br />配置server cookie，以下四项annotation必选。<br />cookie名称(service.beta.kubernetes.io/alicloud-loadbalancer-cookie)只能包含字母、数字、‘_’和‘-’。

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

- **创建负载均衡时，指定主备可用区**

某些region的负载均衡不支持主备可用区，如ap-southeast-5。<br />一旦创建，主备可用区不支持修改。

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

- **使用Pod所在的节点作为后端服务器**

  默认externalTrafficPolicy为Cluster模式，会将集群中所有节点挂载到后端服务器。Local模式仅将Pod所在节点作为后端服务器。

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

- **创建私有网络类型（VPC）的负载均衡**

  创建私有网络类型的负载均衡，以下两个annotation必选。<br />私网负载均衡支持专有网络(vpc)和经典网络(classic)，两者区别参考[实例概述](https://help.aliyun.com/document_detail/85931.html?spm=5176.8009612.101.9.6fb671b3b0DU4I)。

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

- **创建按流量付费的负载均衡**

  仅支持公网类型的负载均衡实例<br />以下两项annotation必选

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

- **创建带健康检查的负载均衡**

##### 设置TCP类型的健康检查
  TCP端口默认开启健康检查，且不支持修改，即service.beta.kubernetes.io/alicloud-loadbalancer-health-check-flag annotation无效。<br />
  设置TCP类型的健康检查，以下所有annotation必选。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
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

##### 设置HTTP类型的健康检查
设置HTTP类型的健康检查，以下所有的annotation必选。

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

- **为负载均衡设置调度算法**

**wrr**（默认值）：权重值越高的后端服务器，被轮询到的次数（概率）也越高。<br />**wlc**：除了根据每台后端服务器设定的权重值来进行轮询，同时还考虑后端服务器的实际负载（即连接数）。当权重值相同时，当前连接数越小的后端服务器被轮询到的次数（概率）也越高。<br />**rr**：按照访问顺序依次将外部请求依序分发到后端服务器。

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

- **创建带有访问控制的负载均衡**

需要先在阿里云控制台上创建一个访问控制并记录acl-id，然后使用如下 annotation 创建一个带有访问控制的负载均衡实例。<br />白名单适合只允许特定IP访问的场景，black黑名单适用于只限制某些特定IP访问的场景。<br />创建带有访问控制的负载均衡，以下三项annotation必选。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-status: "on"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-id: "${YOUR_ACL_ID}"
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

- **为负载均衡指定虚拟交换机**

通过阿里云[专有网络控制台](https://vpc.console.aliyun.com)查询交换机ID，然后使用如下的annotation为负载均衡实例指定虚拟交换机。<br />为负载均衡指定虚拟交换机，以下两项annotation必选。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
   service.beta.kubernetes.io/alicloud-loadbalancer-address-type: "intranet"
   service.beta.kubernetes.io/alicloud-loadbalancer-vswitch-id: "${YOUR_VSWITCH_ID}"
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

- **为负载均衡指定转发端口**

端口转发是指将http端口的请求转发到https端口上。<br />设置端口转发需要先在阿里云控制台上创建一个证书并记录cert-id<br />如需设置端口转发，以下三项annotation必选。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port: "https:443,http:80"
    service.beta.kubernetes.io/alicloud-loadbalancer-cert-id: "${YOUR_CERT_ID}"
    service.beta.kubernetes.io/alicloud-loadbalancer-forward-port: "80:443"
  name: nginx
  namespace: default
spec:
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 443
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

- **为负载均衡添加额外标签**

多个tag以逗号分隔，例如："k1=v1,k2=v2"。

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
  
  
**说明**：注解的内容是区分大小写的，所有注解值均为string类型。

| 注释 | 类型 | 描述 | 默认值 |
| :--- | --- | --- | --- |
| service.beta.kubernetes.io/alicloud-loadbalancer-protocol-port | string | 多个值之间由逗号分隔，比如：https:443,http:80 | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-address-type | string | 取值可以是internet或者intranet | internet |
| service.beta.kubernetes.io/alicloud-loadbalancer-slb-network-type | string | 负载均衡的网络类型，取值：classic或vpc。<br />取值为vpc时，需设置service.beta.kubernetes.io/alicloud-loadbalancer-address-type为intranet。 | classic |
| service.beta.kubernetes.io/alicloud-loadbalancer-charge-type | string | 取值可以是paybytraffic或者paybybandwidth | paybytraffic |
| service.beta.kubernetes.io/alicloud-loadbalancer-id | string | 负载均衡实例的 ID。通过 **service.beta.kubernetes.io/alicloud-loadbalancer-id**指定您已有的SLB，默认情况下，使用已有的负载均衡实例，不会覆盖监听，如要强制覆盖已有监听，请配置**service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners**为true。 | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-backend-label | string | 通过 label 指定 SLB 后端挂载哪些worker节点。 | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-spec | string | 负载均衡实例的规格。可参考：[文档](https://help.aliyun.com/document_detail/27577.html?spm=a2c4g.11186623.2.22.6be5609abM821g#SLB-api-CreateLoadBalancer) | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-persistence-timeout | string | 会话保持时间。<br />仅针对TCP协议的监听，取值：0-3600（秒），默认情况下，取值为0，会话保持关闭。<br />可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener) | 0 |
| service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session | string | 是否开启会话保持。取值：on<br />**说明：** 仅对HTTP和HTTPS协议的监听生效。<br />可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener) | off<br />[](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)<br /> |
| service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type | string | cookie的处理方式。<br />取值：<br />**insert**：植入Cookie。<br />**server**：重写Cookie。<br />**说明：**  <br />- 仅对HTTP和HTTPS协议的监听生效。 <br />- 当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session取值为on时，该参数必选。<br />可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)<br /> | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-cookie-timeout | string | Cookie超时时间。取值：1-86400（秒）<br />**说明：** 当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session为on且service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type为insert时，该参数必选。<br />可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)<br /> | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-cookie | string | 服务器上配置的Cookie名称。长度为1-200个字符，只能包含字母、数字、‘_’和‘-’。<br />**说明：** <br />当service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session为on且service.beta.kubernetes.io/alicloud-loadbalancer-sticky-session-type为server时，该参数必选。<br /><br />参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)<br /> | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-master-zoneid | string | 主后端服务器的可用区ID。 | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-slave-zoneid | string | 备后端服务器的可用区ID。 | 无 |
| externalTrafficPolicy | string | 哪些节点可以作为后端服务器，取值：<br />**Cluster**：使用所有后端节点作为后端服务器。<br /> **Local**：使用Pod所在节点作为后端服务器。 | Cluster |
| service.beta.kubernetes.io/alicloud-loadbalancer-force-override-listeners | string | 绑定已有负载均衡时，是否强制覆盖该SLB的监听。 | false：不覆盖 |
| service.beta.kubernetes.io/alicloud-loadbalancer-bandwidth | string | 负载均衡的带宽，仅适用于公网类型的负载均衡。 | 50 |
| service.beta.kubernetes.io/alicloud-loadbalancer-cert-id | string | 阿里云上的证书 ID。您需要先上传证书。 | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-flag | string | 取值是on&#124;off。TCP监听默认为on且不可更改。HTTP监听默认为off。 | off |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-type | string | 健康检查类型，取值：tcp或http。可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br /> | tcp |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-uri | string | 用于健康检查的URI。**说明：** 当健康检查类型为TCP模式时，无需配置该参数。<br />可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br /> | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-port | string | 健康检查使用的端口。<br />取值：<br />**-520**：默认使用监听配置的后端端口。<br /> **1-65535**：健康检查的后端服务器的端口。<br />可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br /> | 无 |
| service.beta.kubernetes.io/alicloud-loadbalancer-healthy-threshold | string | 健康检查连续成功多少次后，将后端服务器的健康检查状态由fail判定为success。<br />取值：2-10<br />可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br /> | 3 |
| service.beta.kubernetes.io/alicloud-loadbalancer-unhealthy-threshold | string | 健康检查连续失败多少次后，将后端服务器的健康检查状态由success判定为fail。<br />取值：2-10<br />可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br /> | 3 |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval | string | 健康检查的时间间隔。<br /> 取值：1-50（秒）<br /> 可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br /> | 2 |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout | string | 接收来自运行状况检查的响应需要等待的时间,适用于TCP模式。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。<br /> 取值：1-300（秒）<br /> **说明：** 如果service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout的值小于service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值，则service.beta.kubernetes.io/alicloud-loadbalancer-health-check-connect-timeout无效，超时时间为service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值。<br /> 可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)<br /> | 5 |
| service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout | string | 接收来自运行状况检查的响应需要等待的时间，适用于HTTP模式。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。取值：1-300（秒）<br />**说明：** 如果 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout的值小于service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值，则 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-timeout无效，超时时间为 service.beta.kubernetes.io/alicloud-loadbalancer-health-check-interval的值。<br />可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)<br /> | 5 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-domain | string | 用于健康检查的域名。<br />**$_ip**： 后端服务器的私网IP。当指定了IP或该参数未指定时，负载均衡会使用各后端服务器的私网IP当做健康检查使用的域名。<br />**domain**：域名长度为1-80，只能包含字母、数字、点号（.）和连字符（-）。 | 无 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-httpcode | string | 健康检查正常的HTTP状态码，多个状态码用逗号（,）分割。<br />取值：http_2xx（默认值）或http_3xx或http_4xx或http_5xx。 | http_2xx |
| service.beta.kubernetes.io/alicloud-loadbalancer-scheduler | string | 调度算法。取值 wrr或wlc或rr。<br />**wrr**（默认值）：权重值越高的后端服务器，被轮询到的次数（概率）也越高。<br />**wlc**：除了根据每台后端服务器设定的权重值来进行轮询，同时还考虑后端服务器的实际负载（即连接数）。当权重值相同时，当前连接数越小的后端服务器被轮询到的次数（概率）也越高。<br />**rr**：按照访问顺序依次将外部请求依序分发到后端服务器。 | wrr |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-status | string | 是否开启访问控制功能。取值：on或off | off |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-id | string | 监听绑定的访问策略组ID。当AclStatus参数的值为on时，该参数必选。 | 无 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-type | string | 访问控制类型。<br />取值：white或black。<br />**white**： 仅转发来自所选访问控制策略组中设置的IP地址或地址段的请求，白名单适用于应用只允许特定IP访问的场景。设置白名单存在一定业务风险。一旦设名单，就只有白名单中的IP可以访问负载均衡监听。如果开启了白名单访问，但访问策略组中没有添加任何IP，则负载均衡监听会转发全部请求。<br />**black**： 来自所选访问控制策略组中设置的IP地址或地址段的所有请求都不会转发，黑名单适用于应用只限制某些特定IP访问的场景。如果开启了黑名单访问，但访问策略组中没有添加任何IP，则负载均衡监听会转发全部请求。当AclStatus参数的值为on时，该参数必选。 | 无 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-vswitch-id | string | 负载均衡实例所属的VSwitch ID。设置该参数时需同时设置addresstype为intranet。 | 无 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-forward-port | string | 将HTTP请求转发至HTTPS指定端口。取值如80:443 | 无 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags | string | 需要添加的Tag列表，多个标签用逗号分隔。如："k1=v1,k2=v2" | 无 |
