# Cloud Provider 帮助文档（中文版）

# 通过负载均衡（Server Load Balancer）访问服务

您可以使用阿里云负载均衡来访问服务。  
详细信息请参考官方文档：[通过负载均衡（Server Load Balancer）访问服务](https://help.aliyun.com/document_detail/86531.html?spm=5176.10695662.1996646101.searchclickresult.87d74fdf8ZwPdN)

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
- **自动刷新配置**  
  CCM使用声明式API，会在一定条件下自动根据service的配置刷新阿里云负载均衡配置，所有用户自行在SLB控制台上修改的配置均存在被覆盖的风险（使用已有SLB同时不覆盖监听的场景除外），因此不能在SLB控制台手动修改Kubernetes创建并维护的SLB的任何配置，否则有配置丢失的风险。
- 同时支持为serivce指定一个已有的负载均衡，或者让CCM自行创建新的负载均衡。但两种方式在SLB的管理方面存在一些差异。
  **指定已有SLB**  
  - 仅支持复用负载均衡控制台创建的SLB，不支持复用CCM创建的SLB。
  - 需要为Service设置`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id` annotation。
  - SLB配置  
    此时CCM会使用该SLB做为Service的SLB，并根据其他annotation配置SLB，并且自动的为SLB创建多个虚拟服务器组（当集群节点变化的时候，也会同步更新虚拟服务器组里面的节点）。
  - 监听配置  
    是否配置监听取决于`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners: `是否设置为true。 如果设置为false，那么CCM不会为SLB管理任何监听配置。如果设置为true，那么CCM会尝试为SLB更新监听，此时CCM会根据监听名称判断SLB上的监听是否为k8s维护的监听（名字的格式为k8s/Port/ServiceName/Namespace/ClusterID），若Service声明的监听与用户自己管理的监听端口冲突，那么CCM会报错。
  - SLB的删除  
    当Service删除的时候CCM不会删除用户通过id指定的已有SLB。
- **CCM管理的SLB**  
  - CCM会根据service的配置自动的创建配置**SLB**、**监听**、**虚拟服务器组**等资源，所有资源归CCM管理，因此用户不得手动在SLB控制台更改以上资源的配置，否则CCM在下次Reconcile的时候将配置刷回service所声明的配置，造成非用户预期的结果。
  - SLB的删除
    当Service删除的时候CCM会删除该SLB。
- **后端服务器更新**
  - CCM会自动的为该Service对应的SLB刷新后端虚拟服务器组。当Service对应的后端Endpoint发生变化的时候或者集群节点变化的时候都会自动的更新SLB的后端Server。
  - `spec.ExternalTraffic = Cluster`模式的Service，CCM默认会将所有节点挂载到SLB的后端（使用BackendLabel标签配置后端的除外）。由于SLB限制了每个ECS上能够attach的SLB的个数（quota），因此这种方式会快速的消耗该quota,当quota耗尽后，会造成Service Reconcile失败。解决的办法，可以使用Local模式的Service。
  - `spec.ExternalTraffic = Local`模式的Service，CCM默认只会将Service对应的Pod所在的节点加入到SLB后端。这会明显降低quota的消耗速度。同时支持四层源IP保留。
  - 任何情况下CCM不会将Master节点作为SLB的后端。
  - CCM默认不会从SLB后端移除被kubectl drain/cordon的节点。如需移除节点，请设置service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend为on。
  >> 说明 
     如果是v1.9.3.164-g2105d2e-aliyun之前的版本，CCM默认会从SLB后端移除被kubectl drain/cordon的节点。  
- **VPC路由**  
  - 集群中一个节点对应一条路由表项，VPC默认情况下仅支持48条路由表项，如果集群节点数目多于48个，请提工单给VPC产品。 
  >> 说明 您可以在提交工单时，说明需要修改vpc_quota_route_entrys_num参数，用于提升单个路由表可创建的自定义路由条目的数量。 
  - 更多VPC使用限制请参见[使用限制](https://help.aliyun.com/document_detail/27750.html#concept_dyx_jkx_5db)  
  - 专有网络VPC配额查询请参见[专有网络VPC配额管理](https://vpc.console.aliyun.com/quota)  
  
  
- **SLB使用限制**  
  - CCM会为Type=LoadBalancer类型的Service创建SLB。默认情况下一个用户可以保留60个SLB实例，如果需要创建的SLB数量大于60，请提交工单给SLB产品。
  >> 说明 您可以在提交工单时，说明需要修改slb_quota_instances_num参数，用于提高用户可保有的slb实例个数。
  - CCM会根据Service将ECS挂载到SLB后端服务器组中。默认情况下一个ECS实例可挂载的后端服务器组的数量为50个，如果一台ECS需要挂载到更多的后端服务器组中，请提交工单给SLB产品。
  >> 说明 您可以在提交工单时，说明需要修改slb_quota_backendservers_num参数，用于提高同一台服务器可以重复添加为SLB后端服务器的次数。
  - 默认情况下一个SLB实例可以挂载200个后端服务器，如果需要挂载更多的后端服务器，请提交工单给SLB产品。
  >> 说明 您可以在提交工单时，说明需要修改slb_quota_backendservers_num参数，提高每个SLB实例可以挂载的服务器数量。
  - CCM会根据Service中定义的端口创建SLB监听。默认情况下一个SLB实例可以添加50个监听，如需添加更多监听，请提交工单给SLB产品。
  >> 说明 您可以在提交工单时，说明需要修改slb_quota_listeners_num参数，用于提高每个实例可以保有的监听数量。
  - 更多SLB使用限制请参见[使用限制](https://help.aliyun.com/document_detail/32459.html?spm=a2c4g.11174283.3.2.28581192X1CXYW)  
  - 负载均衡SLB配额查询请参见[负载均衡SLB配额管理](https://slbnew.console.aliyun.com/slb/quota)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    >
  
## 通过命令行操作

Step 1. 通过命令行工具创建一个 Nginx 应用。
```bash
root@master # kubectl run nginx -image=registry.aliyuncs.com/acs/netdia:latest
root@master # kubectl get po
NAME                                   READY     STATUS    RESTARTS   AGE
nginx-2721357637-dvwq3                 1/1       Running   1          6s
```

Step 2. 为 Nginx 应用创建阿里云负载均衡服务，指定 `type=LoadBalancer` 来向外网用户暴露 Nginx 服务。
```bash
root@master # kubectl expose deployment nginx --port=80 --target-port=80 --type=LoadBalancer
root@master # kubectl get svc
NAME                  CLUSTER-IP      EXTERNAL-IP      PORT(S)                        AGE
nginx                 172.19.10.209   101.37.192.20   80:31891/TCP                   4s
```

Step 3. 在浏览器中访问 `http://101.37.192.20`，来访问您的 Nginx 服务。

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
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-address-type: "intranet"
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
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port: "http:80"
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

> HTTPS请求会在SLB层解密，然后以HTTP请求的形式发送给后端Pod。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port: "https:443"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cert-id: "${YOUR_CERT_ID}"
  name: nginx
  namespace: default
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 80
  selector:
    run: nginx
  type: LoadBalancer
```

- **指定负载均衡规格**

  负载均衡规格可参考[文档](https://help.aliyun.com/document_detail/27577.html?spm=a2c4g.11186623.2.22.6be5609abM821g#SLB-api-CreateLoadBalancer)。
  通过该参数可以创建指定规格的SLB，或者更新已有SLB的规格。  
  注意：如果您通过SLB控制台修改SLB规格，将会存在被CCM修改回原规格的风险，请谨慎操作。    
  
```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-spec: "slb.s1.small"
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

  默认情况下，使用已有的负载均衡实例不会覆盖监听，如需强制覆盖已有监听，请配置`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners`为true。  
  使用已有的负载均衡暂不支持添加额外标签（annotation: service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags）  

>> 复用已有的负载均衡默认不覆盖已有监听，出于以下两点原因：  
  1）如果已有负载均衡的监听上绑定了业务，强制覆盖可能会引发业务中断  
  2）由于CCM目前支持的后端配置有限，无法处理一些复杂配置。如果有复杂的后端配置需求，用户可以在不覆盖监听的情况下，通过控制台自行配置监听。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id: "${YOUR_LOADBALACER_ID}"
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
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id: "${YOUR_LOADBALACER_ID}"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners: "true"
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

- **使用指定label的worker节点作为后端服务器**

  多个label以逗号分隔。例如："k1=v1,k2=v2"。多个label之间是`and`的关系。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-backend-label: "failure-domain.beta.kubernetes.io/zone=ap-southeast-5a"
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

  `service.beta.kubernetes.io/alibaba-cloud-loadbalancer-persistence-time`仅对TCP协议的监听生效。  
  如果负载均衡实例配置了多个TCP协议的监听端口，则默认将该配置应用到所有TCP协议的监听端口。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-persistence-timeout: "1800"
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

  仅支持HTTP及HTTPS协议的负载均衡实例。  
  如果配置了多个HTTP或者HTTPS的监听端口，该会话保持默认应用到所有HTTP和HTTPS监听端口。  
  配置insert cookie，以下四项annotation必选。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session: "on"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session-type: "insert"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cookie-timeout: "1800"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port: "http:80"
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

  仅支持HTTP及HTTPS协议的负载均衡实例。  
  如果配置了多个HTTP或者HTTPS的监听端口，该会话保持默认应用到所有HTTP和HTTPS监听端口。  
  配置server cookie，以下四项annotation必选。  
  cookie名称(service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cookie)只能包含字母、数字、‘_’和‘-’。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session: "on"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session-type: "server"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cookie: "${YOUR_COOKIE}"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port: "http:80"
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

  某些region的负载均衡不支持主备可用区，如ap-southeast-5。  一旦创建，主备可用区不支持修改。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-master-zoneid: "ap-southeast-5a"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-slave-zoneid: "ap-southeast-5a"
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

  默认externalTrafficPolicy为Cluster模式，会将集群中所有节点挂载到后端服务器。  
  Local模式（即设置externalTrafficPolicy: Local）仅将Pod所在节点作为后端服务器。
  Local模式需要设置调度策略为加权轮询wrr。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-scheduler: "wrr"
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

  创建私有网络类型的负载均衡，以下两个annotation必选。  
  私网负载均衡支持专有网络(vpc)和经典网络(classic)，两者区别参考[实例概述](https://help.aliyun.com/document_detail/85931.html?spm=5176.8009612.101.9.6fb671b3b0DU4I)。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-address-type: "intranet"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-network-type: "alibaba"
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

- **创建按带宽计费的负载均衡**

  `service.beta.kubernetes.io/alibaba-cloud-loadbalancer-bandwidth`为带宽峰值  
  
  仅适用于公网类型的负载均衡实例  
  
  其他限制请参考[修改公网负载均衡实例的计费方式](https://help.aliyun.com/document_detail/27578.html?spm=a2c4g.11186623.6.701.22482bcdAgzI6s)  
  
  以下两项annotation必选

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-charge-type: "paybybandwidth"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-bandwidth: "45"
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
  TCP端口默认开启健康检查，且不支持修改，即service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-flag annotation无效。  
  设置TCP类型的健康检查，以下所有annotation必选。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-type: "tcp"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-connect-timeout: "8"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-healthy-threshold: "4"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-unhealthy-threshold: "4"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-interval: "3"
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
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-flag: "on"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-type: "http"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-uri: "/test/index.html"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-healthy-threshold: "4"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-unhealthy-threshold: "4"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-timeout: "10"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-interval: "3"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port: "http:80"
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

  **rr**（默认值）：按照访问顺序依次将外部请求依序分发到后端服务器。
  **wrr**：权重值越高的后端服务器，被轮询到的次数（概率）也越高。  
  **wlc**：除了根据每台后端服务器设定的权重值来进行轮询，同时还考虑后端服务器的实际负载（即连接数）。当权重值相同时，当前连接数越小的后端服务器被轮询到的次数（概率）也越高。  
  

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-scheduler: "wlc"
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

  需要先在阿里云控制台上创建一个访问控制并记录acl-id，然后使用如下annotation创建一个带有访问控制的负载均衡实例。  
  白名单适合只允许特定IP访问的场景，black黑名单适用于只限制某些特定IP访问的场景。  
  创建带有访问控制的负载均衡，以下三项annotation必选。

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

  通过阿里云[专有网络控制台](https://vpc.console.aliyun.com)查询交换机ID，然后使用如下的annotation为负载均衡实例指定虚拟交换机。   
  虚拟交换机必须与Kubernetes集群属于同一个VPC。  
  为负载均衡指定虚拟交换机，以下两项annotation必选。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
   service.beta.kubernetes.io/alibaba-cloud-loadbalancer-address-type: "intranet"
   service.beta.kubernetes.io/alibaba-cloud-loadbalancer-vswitch-id: "${YOUR_VSWITCH_ID}"
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

  端口转发是指将http端口的请求转发到https端口上。  
  设置端口转发需要先在阿里云控制台上创建一个证书并记录cert-id。  
  如需设置端口转发，以下三项annotation必选。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port: "https:443,http:80"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cert-id: "${YOUR_CERT_ID}"
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-forward-port: "80:443"
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

- **移除slb后端不可调度节点**

  默认不移除不可调度节点。  
  设置annotation:`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend`为"on"，会将不可调度节点从slb后端移除。

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend: "on"
  name: nginx
spec:
  externalTrafficPolicy: Local
  ports:
  - name: http
    port: 30080
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: LoadBalancer
```

- **将pod ENI挂载到slb后端**  

  在[terway](https://www.alibabacloud.com/help/zh/doc-detail/97467.html?spm=a2c5t.10695662.1996646101.searchclickresult.2304c302tiORcM)网络模式下,通过设定annotation`service.beta.kubernetes.io/backend-type`为eni，可将pod ENI直接挂载到slb后端，提升网络转发性能。

```yaml
  apiVersion: v1
  kind: Service
  metadata:
    annotations:
      service.beta.kubernetes.io/backend-type: "eni"
    name: nginx
  spec:
    ports:
    - name: http
      port: 30080
      protocol: TCP
      targetPort: 80
    selector:
      app: nginx
    type: LoadBalancer
```

- **创建IPv6类型的负载均衡**  

  集群的kube-proxy代理模式需要是IPVS。  
  生成的IPv6地址仅可在支持IPv6的环境中访问。  
  创建后IP类型不可更改。  
```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/alibaba-cloud-loadbalancer-ip-version: "ipv6"
  name: nginx
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: LoadBalancer
```
- **SLB后端混合挂载ECS和弹性网卡ENI** 

  集群中部署虚拟节点，详情请参见[在ACK集群中部署虚拟节点Addon](https://help.aliyun.com/document_detail/118970.html?spm=a2c4g.11186623.6.873.507e76fbzlrvZ5)。  
  应用Pod同时运行在ECS和虚拟节点上，详情请参见[调度Pod到虚拟节点](https://help.aliyun.com/document_detail/118970.html?spm=a2c4g.11186623.6.873.507e76fbzlrvZ5)。   
  SLB创建成功后，可在SLB后端同时看到ECS和弹性网卡ENI。  
  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: nginx
  spec:
    ports:
    - port: 80
      protocol: TCP
      targetPort: 80
    selector:
      app: nginx
    type: LoadBalancer
  ```
  
- 为负载均衡开启删除保护

  默认开启删除保护。

  > 注意：对于LoadBalancer类型的service创建的负载均衡，如手动在SLB控制台开启了删除保护，仍可通过`kubectl delete svc xxx`的方式删除service关联的负载均衡。

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    annotations:
      service.beta.kubernetes.io/alibaba-cloud-loadbalancer-delete-protection: "on"
    name: nginx
  spec:
    externalTrafficPolicy: Local
    ports:
    - port: 80
      protocol: TCP
      targetPort: 80
    selector:
      app: nginx
    type: LoadBalancer
  ```

- 为负载均衡开启配置修改保护

  默认开启配置修改保护。

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    annotations:
      service.beta.kubernetes.io/alibaba-cloud-loadbalancer-modification-protection: "ConsoleProtection"
    name: nginx
  spec:
    externalTrafficPolicy: Local
    ports:
    - port: 80
      protocol: TCP
      targetPort: 80
    selector:
      app: nginx
    type: LoadBalancer
  ```

- 指定负载均衡名称

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    annotations:
      service.beta.kubernetes.io/alibaba-cloud-loadbalancer-name: "your-svc-name"
    name: nginx
  spec:
    externalTrafficPolicy: Local
    ports:
    - port: 80
      protocol: TCP
      targetPort: 80
    selector:
      app: nginx
    type: LoadBalancer
  ```

- 指定负载均衡所属的资源组

  通过阿里云[资源管理平台](https://resourcemanager.console.aliyun.com/)查询资源组ID，然后使用如下的annotation为负载均衡实例指定资源组。   

  资源组ID创建后不可修改。

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    annotations:
      service.beta.kubernetes.io/alibaba-cloud-loadbalancer-resource-group-id: "rg-xxxx"
    name: nginx
  spec:
    externalTrafficPolicy: Local
    ports:
    - port: 80
      protocol: TCP
      targetPort: 80
    selector:
      app: nginx
    type: LoadBalancer
  ```

**说明**

- 注解的内容区分大小写。
- 自2019年9月11日起，annotation字段alicloud更新为alibaba-cloud。  
  例如：  
  &emsp;&emsp;更新前：service.beta.kubernetes.io/alicloud-loadbalancer-id  
  &emsp;&emsp;更新后：service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id  
  系统将继续兼容alicloud的写法，用户无需做任何修改，敬请注意。  

| 注释 | 类型 | 描述 | 默认值 | 支持的版本 |
| :--- | --- | --- | --- | --- |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port | string | 多个值之间由逗号分隔，比如：https:443,http:80 | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-address-type | string | 取值可以是internet或者intranet | internet | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-slb-network-type | string | 负载均衡的网络类型，取值：classic或vpc。  取值为vpc时，需设置service.beta.kubernetes.io/alibaba-cloud-loadbalancer-address-type为intranet。 | classic | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-charge-type | string | 取值可以是paybytraffic或者paybybandwidth | paybytraffic | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id | string | 负载均衡实例的 ID。通过 **service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id**指定您已有的SLB，默认情况下，使用已有的负载均衡实例，不会覆盖监听，如要强制覆盖已有监听，请配置**service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners**为true。 | 无 | v1.9.3.59-ge3bc999-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-backend-label | string | 通过 label 指定 SLB 后端挂载哪些worker节点。 | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-spec | string | 负载均衡实例的规格。可参考：[文档](https://help.aliyun.com/document_detail/27577.html?spm=a2c4g.11186623.2.22.6be5609abM821g#SLB-api-CreateLoadBalancer) | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-persistence-timeout | string | 会话保持时间。  仅针对TCP协议的监听，取值：0-3600（秒），默认情况下，取值为0，会话保持关闭。  可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener) | 0 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session | string | 是否开启会话保持。取值：on  **说明：** 仅对HTTP和HTTPS协议的监听生效。  可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener) | off  [](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)   | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session-type | string | cookie的处理方式。  取值：  **insert**：植入Cookie。  **server**：重写Cookie。  **说明：**    - 仅对HTTP和HTTPS协议的监听生效。   - 当service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session取值为on时，该参数必选。  可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)   | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cookie-timeout | string | Cookie超时时间。取值：1-86400（秒）  **说明：** 当service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session为on且service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session-type为insert时，该参数必选。  可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)   | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cookie | string | 服务器上配置的Cookie名称。长度为1-200个字符，只能包含字母、数字、‘_’和‘-’。  **说明：**   当service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session为on且service.beta.kubernetes.io/alibaba-cloud-loadbalancer-sticky-session-type为server时，该参数必选。    参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)和[CreateLoadBalancerHTTPSListener](https://help.aliyun.com/document_detail/27593.html?spm=a2c4g.11186623.2.25.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPSListener)   | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-master-zoneid | string | 主后端服务器的可用区ID。 | 无 | v1.9.3.10-gfb99107-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-slave-zoneid | string | 备后端服务器的可用区ID。 | 无 | v1.9.3.10-gfb99107-aliyun及以上版本 |
| externalTrafficPolicy | string | 哪些节点可以作为后端服务器，取值：  **Cluster**：使用所有后端节点作为后端服务器。   **Local**：使用Pod所在节点作为后端服务器。 | Cluster | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners | string | 绑定已有负载均衡时，是否强制覆盖该SLB的监听。 | false：不覆盖 | v1.9.3.59-ge3bc999-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-bandwidth | string | 负载均衡的带宽，仅适用于公网类型的负载均衡。 | 50 | v1.9.3.10-gfb99107-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cert-id | string | 阿里云上的证书 ID。您需要先上传证书。 | 无 | v1.9.3.164-g2105d2e-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-flag | string | 取值是on&#124;off。TCP监听默认为on且不可更改。HTTP监听默认为off。 | off | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-type | string | 健康检查类型，取值：tcp或http。可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)   | tcp | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-uri | string | 用于健康检查的URI。**说明：** 当健康检查类型为TCP模式时，无需配置该参数。  可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)   | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-connect-port | string | 健康检查使用的端口。  取值：  **-520**：默认使用监听配置的后端端口。   **1-65535**：健康检查的后端服务器的端口。  可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)   | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-healthy-threshold | string | 健康检查连续成功多少次后，将后端服务器的健康检查状态由fail判定为success。  取值：2-10  可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)   | 3 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-unhealthy-threshold | string | 健康检查连续失败多少次后，将后端服务器的健康检查状态由success判定为fail。  取值：2-10  可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)   | 3 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-interval | string | 健康检查的时间间隔。   取值：1-50（秒）   可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)   | 2 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-connect-timeout | string | 接收来自运行状况检查的响应需要等待的时间,适用于TCP模式。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。   取值：1-300（秒）   **说明：** 如果service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-connect-timeout的值小于service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-interval的值，则service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-connect-timeout无效，超时时间为service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-interval的值。   可参考：[CreateLoadBalancerTCPListener](https://help.aliyun.com/document_detail/27594.html?spm=a2c4g.11186623.2.23.16bb609awRQFbk#slb-api-CreateLoadBalancerTCPListener)   | 5 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-timeout | string | 接收来自运行状况检查的响应需要等待的时间，适用于HTTP模式。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。取值：1-300（秒）  **说明：** 如果 service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-timeout的值小于service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-interval的值，则 service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-timeout无效，超时时间为 service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-interval的值。  可参考：[CreateLoadBalancerHTTPListener](https://help.aliyun.com/document_detail/27592.html?spm=a2c4g.11186623.2.24.16bb609awRQFbk#slb-api-CreateLoadBalancerHTTPListener)   | 5 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-domain | string | 用于健康检查的域名。  **$_ip**： 后端服务器的私网IP。当指定了IP或该参数未指定时，负载均衡会使用各后端服务器的私网IP当做健康检查使用的域名。  **domain**：域名长度为1-80，只能包含字母、数字、点号（.）和连字符（-）。 | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-health-check-httpcode | string | 健康检查正常的HTTP状态码，多个状态码用逗号（,）分割。  取值：http_2xx（默认值）或http_3xx或http_4xx或http_5xx。 | http_2xx | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-scheduler | string | 调度算法。取值 wrr或wlc或rr。  **wrr**：权重值越高的后端服务器，被轮询到的次数（概率）也越高。  **wlc**：除了根据每台后端服务器设定的权重值来进行轮询，同时还考虑后端服务器的实际负载（即连接数）。当权重值相同时，当前连接数越小的后端服务器被轮询到的次数（概率）也越高。  **rr**（默认值）：按照访问顺序依次将外部请求依序分发到后端服务器。 | rr | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-status | string | 是否开启访问控制功能。取值：on或off | off | v1.9.3.164-g2105d2e-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-id | string | 监听绑定的访问策略组ID。当AclStatus参数的值为on时，该参数必选。 | 无 | v1.9.3.164-g2105d2e-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-acl-type | string | 访问控制类型。  取值：white或black。  **white**： 仅转发来自所选访问控制策略组中设置的IP地址或地址段的请求，白名单适用于应用只允许特定IP访问的场景。设置白名单存在一定业务风险。一旦设名单，就只有白名单中的IP可以访问负载均衡监听。如果开启了白名单访问，但访问策略组中没有添加任何IP，则负载均衡监听会转发全部请求。  **black**： 来自所选访问控制策略组中设置的IP地址或地址段的所有请求都不会转发，黑名单适用于应用只限制某些特定IP访问的场景。如果开启了黑名单访问，但访问策略组中没有添加任何IP，则负载均衡监听会转发全部请求。当AclStatus参数的值为on时，该参数必选。 | 无 | v1.9.3.164-g2105d2e-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-vswitch-id | string | 负载均衡实例所属的VSwitch ID。设置该参数时需同时设置addresstype为intranet。 | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-forward-port | string | 将HTTP请求转发至HTTPS指定端口。取值如80:443 | 无 | v1.9.3.164-g2105d2e-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags | string | 需要添加的Tag列表，多个标签用逗号分隔。如："k1=v1,k2=v2" | 无 | v1.9.3及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend | string | 从slb后端移除SchedulingDisabled Node。取值：on或off | off | v1.9.3.164-g2105d2e-aliyun及以上版本 |
| service.beta.kubernetes.io/backend-type | string | 支持在terway eni网络模式下,通过设定该参数为"eni"，可将pod直接挂载到slb后端，提升网络转发性能。取值：eni | 无 | v1.9.3.164-g2105d2e-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-ip-version | string | 负载均衡实例的IP版本，取值：ipv4或ipv6 | ipv4 | v1.9.3.220-g24b1885-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-delete-protection | string | 负载均衡删除保护，取值：on或off | on | v1.9.3.304-g1f42462-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-modification-protection | string | 负载均衡配置修改保护，取值：ConsoleProtection或NonProtection | ConsoleProtection | v1.9.3.304-g1f42462-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-resource-group-id | string | 负载均衡所属资源组ID | 无 | v1.9.3.304-g1f42462-aliyun及以上版本 |
| service.beta.kubernetes.io/alibaba-cloud-loadbalancer-name | string | 负载均衡实例名称 | 无                                                           | v1.9.3.304-g1f42462-aliyun及以上版本 |
