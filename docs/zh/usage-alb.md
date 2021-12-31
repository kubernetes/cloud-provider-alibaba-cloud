# Cloud Provider 帮助文档（中文版）

# 通过应用型负载均衡（Application Load Balancer）Ingress 访问服务

ALB Ingress基于阿里云应用型负载均衡ALB（Application Load Balancer） 实现的Ingress服务，适用于有明显波峰波谷的业务场景。本文介绍如何在ACK集群中使用ALB Ingress访问服务。
详细信息请参考官方文档：[通过ALB Ingress访问服务](https://help.aliyun.com/document_detail/314614.html)

## 背景信息

Ingress是允许访问集群内Service的规则集合，您可以通过配置转发规则，实现不同URL访问集群内不同的Service。但传统的Nginx Ingress或者四层SLB Ingress，已无法满足云原生应用服务对复杂业务路由、多种应用层协议（例如：QUIC等）、大规模七层流量能力的需求。

## 应用型负载均衡相关信息

- 应用型负载均衡介绍参见[什么是应用型负载均衡ALB](https://help.aliyun.com/document_detail/197202.html)
- 应用型负载均衡版本功能对比和使用限制参见[版本功能对比和使用限制](https://help.aliyun.com/document_detail/197204.html?spm=5176.20310575.help.dexternal.13e31eb9CcVMMQ)
- 应用型负载均衡支持的地域与可用区参见[支持的地域与可用区](https://help.aliyun.com/document_detail/258300.html)

## 配置Albconfig

Albconfig是由ALB Ingress Controller提供的CRD资源，ALB Ingress Controller使用Albconfig来配置ALB实例和监听。本节介绍如何创建、修改Albconfig以及开启日志服务等操作。

### 背景信息
ALB Ingress Controller通过API Server获取Ingress资源的变化，动态地生成Albconfig，然后依次创建ALB实例、监听、路由转发规则以及后端服务器组。Kubernetes中Service、Ingress与Albconfig有着以下关系：

- Service是后端真实服务的抽象，一个Service可以代表多个相同的后端服务。
- Ingress是反向代理规则，用来规定HTTP/HTTPS请求应该被转发到哪个Service上。例如：根据请求中不同的Host和URL路径，让请求转发到不同的Service上。
- Albconfig是在ALB Ingress Controller提供的CRD资源，使用ALBConfig CRD来配置ALB实例和监听。一个Albconfig对应一个ALB实例。
- 一个Albconfig对应一个ALB实例，如果一个ALB实例配置多个转发规则，那么一个Albconfig则对应多个Ingress，所以Albconfig与Ingress是一对多的对应关系。

### 创建Albconfig

创建Ingress时，会默认在kube-system命名空间下创建名称为default的Albconfig，无需您手动创建。

1. 部署以下模板，创建Ingress和Albconfig。
   ```yaml
   apiVersion: networking.k8s.io/v1beta1
   kind: Ingress
   metadata:
     name: cafe-ingress
     annotations:
       kubernetes.io/ingress.class: alb
       alb.ingress.kubernetes.io/address-type: internet
       alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"
   spec:
     rules:
     - http:
         paths:
         # 配置Context Path
         - path: /tea
           backend:
             serviceName: tea-svc
             servicePort: 80
         # 配置Context Path
         - path: /coffee
           backend:
             serviceName: coffee-svc
             servicePort: 80
   ```
2. 执行以下命令，查看Albconfig名称。
    ```bash
    kubectl -n kube-system get albconfig
    ```
   预期输出：
    ```bash
    NAME      AGE
    default   87m
    ```
   Albconfig默认配置的内容如下：
   ```yaml
   apiVersion: alibabacloud.com/v1
   kind: AlbConfig
   metadata:
     name: default                      #Albconfig名称。
     namespace: kube-system             #Albconfig所属命名空间。
   spec:
     config:
       accessLogConfig:
         logProject: ""
         logStore: ""
       addressAllocatedMode: Dynamic
       addressType: Internet
       billingConfig:
         internetBandwidth: 0
         internetChargeType: ""
         payType: PostPay
       deletionProtectionEnabled: true
       edition: Standard
       forceOverride: false
       zoneMappings:
       - vSwitchId: vsw-wz92lvykqj1siwvif****        #Albconfig的vSwitch，Albconfig需要配置两个vSwitch。
       - vSwitchId: vsw-wz9mnucx78c7i6iog****        #Albconfig的vSwitch。
   status:
     loadBalancer:
       dnsname: alb-s2em8fr9debkg5****.cn-shenzhen.alb.aliyuncs.com
       id: alb-s2em8fr9debkg5****
   ```
### 修改Albconfig的名称
如果您需要修改Albconfig的名称，可以执行以下命令。保存之后，新名称自动生效。
```bash
kubectl -n kube-system edit albconfig default
...
  spec:
    config:
      name: basic   #输入修改后的名称。
...
```
### 修改Albconfig的vSwitch配置
如果您需要修改Albconfig的vSwitch配置。保存之后，新配置自动生效。
```bash
kubectl -n kube-system edit albconfig default
...
  zoneMappings:
    - vSwitchId: vsw-wz92lvykqj1siwvif****
    - vSwitchId: vsw-wz9mnucx78c7i6iog****
...
```
### 开启日志服务访问日志
如果您希望ALB Ingress能够收集访问日志Access Log，则只需要在AlbConfig中指定logProject和logStore。
```yaml
apiVersion: alibabacloud.com/v1
kind: AlbConfig
metadata:
  name: default
  namespace: kube-system
spec:
  config:
    accessLogConfig:
      logProject: "k8s-log-xz92lvykqj1siwvif****"
      logStore: "alb_xxx"
    ...
```
```
说明: logStore命名需要以alb_开头，若指定logStore不存在，系统则会自动创建。保存命令之后，可以在日志服务控制台，单击目标Logstore，查看收集的访问日志。
```
### 删除ALB实例
一个ALB实例对应一个Albconfig， 因此可以通过删除Albconfig实现删除ALB实例，但前提是先需要删除Albconfig关联的所有Ingress。
```bash
kubectl -n kube-system delete albconfig default
```
default可以替换为您实际需要删除的Albconfig。

## 通过ALB Ingress访问服务

本节介绍如何在ACK集群中使用ALB Ingress访问服务。

### 注意事项

如果您使用的是Flannel网络插件，则ALB Ingress后端Service服务仅支持NodePort和LoadBalancer类型。

### 步骤一：部署服务

1. 创建并拷贝以下内容到cafe-service.yaml文件中，用于部署两个名称分别为coffee和tea的Deployment，以及两个名称分别为coffee和tea的Service。
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: coffee
   spec:
     replicas: 2
     selector:
       matchLabels:
         app: coffee
     template:
       metadata:
         labels:
           app: coffee
       spec:
         containers:
         - name: coffee
           image: registry.cn-hangzhou.aliyuncs.com/acs-sample/nginxdemos:latest
           ports:
           - containerPort: 80
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: coffee-svc
   spec:
     ports:
     - port: 80
       targetPort: 80
       protocol: TCP
     selector:
       app: coffee
     type: NodePort
   ---
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: tea
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: tea
     template:
       metadata:
         labels:
           app: tea
       spec:
         containers:
         - name: tea
           image: registry.cn-hangzhou.aliyuncs.com/acs-sample/nginxdemos:latest
           ports:
           - containerPort: 80
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: tea-svc
     labels:
   spec:
     ports:
     - port: 80
       targetPort: 80
       protocol: TCP
     selector:
       app: tea
     type: NodePort
   ```
2. 执行以下命令，部署两个Deployment和两个Service。
    ```bash
    kubectl apply -f cafe-service.yaml
    ```
   预期输出：
    ```bash
    deployment "coffee" created
    service "coffee-svc" created
    deployment "tea" created
    service "tea-svc" created
    ```
3. 执行以下命令，查看服务状态。
    ```bash
    kubectl get svc,deploy
    ```
   预期输出：
    ```bash
    NAME             TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
    coffee-svc   NodePort    172.16.231.169   <none>        80:31124/TCP   6s
    tea-svc      NodePort    172.16.38.182    <none>        80:32174/TCP   5s
    NAME            DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
    deploy/coffee   2         2         2            2           1m
    deploy/tea      1         1         1            1           1m
    ```

### 步骤二：配置Ingress

1. 创建并拷贝以下内容到cafe-ingress.yaml文件中。
    ```yaml
    apiVersion: networking.k8s.io/v1beta1
    kind: Ingress
    metadata:
    name: cafe-ingress
    annotations:
        kubernetes.io/ingress.class: alb
        alb.ingress.kubernetes.io/name: ingress_test_base
        alb.ingress.kubernetes.io/address-type: internet
        alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"
    spec:
    rules:
    - http:
        paths:
        # 配置Context Path。
        - path: /tea
            backend:
            serviceName: tea-svc
            servicePort: 80
        # 配置Context Path。
        - path: /coffee
            backend:
            serviceName: coffee-svc
            servicePort: 80
    ```
   相关参数解释如下表所示：
   <table class="table">
   <thead class="thead">
      <tr>
         <th class="entry">参数</th>
         <th class="entry">说明</th>
      </tr>
   </thead>
   <tbody class="tbody">
      <tr>
         <td class="entry">（可选）<span style='font-weight:700'>alb.ingress.kubernetes.io/name</span></td>
         <td class="entry">表示ALB实例名称。</td>
      </tr>
      <tr>
         <td class="entry">（可选）<span style='font-weight:700'>alb.ingress.kubernetes.io/address-type</span></td>
         <td class="entry">表示负载均衡的地址类型。取值如下：
            <ul class="ul">
               <li class="li">Internet（默认值）：负载均衡具有公网IP地址，DNS域名被解析到公网IP，因此可以在公网环境访问。</li>
               <li class="li">Intranet：负载均衡只有私网IP地址，DNS域名被解析到私网IP，因此只能被负载均衡所在VPC的内网环境访问。</li>
            </ul>
         </td>
      </tr>
      <tr>
         <td class="entry">（必选）<span style='font-weight:700'>alb.ingress.kubernetes.io/vswitch-ids</span></td>
         <td class="entry">用于设置ALB Ingress交换机ID，您需要至少指定两个不同可用区交换机ID。关于ALB Ingress支持的地域与可用区，请参见<span><a title="本文介绍支持的地域（Region）与可用区AZ（Availability Zone）。"  href="https://help.aliyun.com/document_detail/258300.htm">支持的地域与可用区</a></span>。
         </td>
      </tr>
   </tbody>
   </table>

2. 执行以下命令，配置coffee和tea服务对外暴露的域名和path路径。
    ```bash
    kubectl apply -f cafe-ingress.yaml
    ```
   预期输出：
    ```bash
    ingress "cafe-ingress" created
    ```
3. 执行以下命令获取ALB实例IP地址。
    ```bash
    kubectl get ing
    ```
   预期输出：
    ```
    NAME           CLASS    HOSTS   ADDRESS                                               PORTS   AGE
    cafe-ingress   <none>   *       alb-m551oo2zn63yov****.cn-hangzhou.alb.aliyuncs.com   80      50s
    ```

### 步骤三：访问服务
- 利用获取的ALB实例IP地址访问coffee服务：
    ```bash
    curl http://alb-m551oo2zn63yov****.cn-hangzhou.alb.aliyuncs.com/coffee
    ```
- 利用获取的ALB实例IP地址访问tea服务：
    ```bash
    curl http://alb-m551oo2zn63yov****.cn-hangzhou.alb.aliyuncs.com/tea
    ```

## ALB Ingress服务高级用法

在Kubernetes集群中，ALB Ingress对集群服务（Service）中外部可访问的API对象进行管理，提供七层负载均衡能力。本文介绍如何使用ALB Ingress将来自不同域名或URL路径的请求转发给不同的后端服务器组、将HTTP访问重定向至HTTPS及实现灰度发布等功能。

### 基于域名转发请求
通过以下命令创建一个简单的Ingress，根据指定的正常域名或空域名转发请求。
- 基于正常域名转发请求的示例如下：
  1. 部署以下模板，分别创建Service、Deployment和Ingress，将访问请求通过Ingress的域名转发至Service。
      ```yaml
      apiVersion: v1
      kind: Service
      metadata:
        name: demo-service
        namespace: default
      spec:
        ports:
          - name: port1
            port: 80
            protocol: TCP
            targetPort: 8080
        selector:
          app: demo
        sessionAffinity: None
        type: NodePort 
      ---
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: demo
        namespace: default
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: demo
        template:
          metadata:
            labels:
              app: demo
          spec:
            containers:
              - image: registry.cn-hangzhou.aliyuncs.com/alb-sample/cafe:v1
                imagePullPolicy: IfNotPresent
                name: demo
                ports:
                  - containerPort: 8080
                    protocol: TCP
      ---
      apiVersion: networking.k8s.io/v1beta1
      kind: Ingress
      metadata:
        annotations:
          alb.ingress.kubernetes.io/address-type: internet
          alb.ingress.kubernetes.io/vswitch-ids: "vsw-2zeqgkyib34gw1fxs****,vsw-2zefv5qwao4przzlo****"
          kubernetes.io/ingress.class: alb
        name: demo
        namespace: default
      spec:
        rules:
          - host: demo.domain.ingress.top
            http:
              paths:
                - backend:
                    serviceName: demo-service
                    servicePort: 80
                  path: /hello
                  pathType: ImplementationSpecific
      ```
  2. 执行以下命令，通过指定的正常域名访问服务。

     替换ADDRESS为ALB实例对应的域名地址，可通过kubectl get ing获取。
     ```bash
     curl -H "host: demo.domain.ingress.top" <ADDRESS>/hello
     ```
     预期输出
     ```bash
     {"hello":"coffee"}
     ```

- 基于空域名转发请求的示例如下：
  1. 部署以下模板，创建Ingress。
      ```yaml
      apiVersion: networking.k8s.io/v1beta1
      kind: Ingress
      metadata:
          annotations:
          alb.ingress.kubernetes.io/address-type: internet
          alb.ingress.kubernetes.io/vswitch-ids: "vsw-2zeqgkyib34gw1fxs****,vsw-2zefv5qwao4przzlo****"
          kubernetes.io/ingress.class: alb
          name: demo
          namespace: default
      spec:
          rules:
          - host: ""
              http:
              paths:
                  - backend:
                      serviceName: demo-service
                      servicePort: 80
                  path: /hello
                  pathType: ImplementationSpecific
      ```
  2. 执行以下命令，通过空域名访问服务。

     替换ADDRESS为ALB实例对应的域名地址，可通过kubectl get ing获取。
      ```bash
      curl <ADDRESS>/hello
      ```
     预期输出：
      ```bash
      {"hello":"coffee"}
      ```

### 基于URL路径转发请求

ALB Ingress支持按照URL转发请求，可以通过pathType字段设置不同的URL匹配策略。pathType支持Exact、ImplementationSpecific和Prefix三种匹配方式。

三种匹配方式的示例如下：
- Exact：以区分大小写的方式精确匹配URL路径。
  1. 部署以下模板，创建Ingress。
      ```yaml
      apiVersion: networking.k8s.io/v1beta1
      kind: Ingress
      metadata:
        annotations:
          alb.ingress.kubernetes.io/vswitch-ids: "vsw-2zeqgkyib34gw1fxs****,vsw-2zefv5qwao4przzlo****"
          kubernetes.io/ingress.class: alb
        name: demo-path
        namespace: default
      spec:
        rules:
          - http:
              paths:
              - path: /hello
                backend:
                  serviceName: demo-service
                  servicePort: 80
                pathType: Exact
      ```
  2. 执行以下命令，访问服务。

           替换ADDRESS为ALB实例对应的域名地址，可通过kubectl get ing获取。
            ```bash
            curl <ADDRESS>/hello
            ```
           预期输出：
            ```bash
            {"hello":"coffee"}
            ```
- ImplementationSpecific：缺省。在ALB Ingress中与Exact做相同处理，但两者Ingress Controller的实现方式不一样。
  1. 部署以下模板，创建Ingress。
      ```yaml
      apiVersion: networking.k8s.io/v1beta1
      kind: Ingress
      metadata:
        annotations:
          alb.ingress.kubernetes.io/address-type: internet
          alb.ingress.kubernetes.io/vswitch-ids: "vsw-2zeqgkyib34gw1fxs****,vsw-2zefv5qwao4przzlo****"
          kubernetes.io/ingress.class: alb
        name: demo-path
        namespace: default
      spec:
        rules:
          - http:
              paths:
              - path: /hello
                backend:
                  serviceName: demo-service
                  servicePort: 80
                pathType: ImplementationSpecific
      ```
  2. 执行以下命令，访问服务。

     替换ADDRESS为ALB实例对应的域名地址，可通过kubectl get ing获取。
      ```bash
      curl <ADDRESS>/hello
      ```
     预期输出：
      ```bash
      {"hello":"coffee"}
      ```
- Prefix：以/分隔的URL路径进行前缀匹配。匹配区分大小写，并且对路径中的元素逐个完成匹配。
  1. 部署以下模板，创建Ingress。
      ```yaml
      apiVersion: networking.k8s.io/v1beta1
      kind: Ingress
      metadata:
        annotations:
          alb.ingress.kubernetes.io/address-type: internet
          alb.ingress.kubernetes.io/vswitch-ids: "vsw-2zeqgkyib34gw1fxs****,vsw-2zefv5qwao4przzlo****"
          kubernetes.io/ingress.class: alb
        name: demo-path-prefix
        namespace: default
      spec:
        rules:
          - http:
              paths:
              - path: /
                backend:
                  serviceName: demo-service
                  servicePort: 80
                pathType: Prefix
      ```
  2. 执行以下命令，访问服务。

     替换ADDRESS为ALB实例对应的域名地址，可通过kubectl get ing获取。
      ```bash
      curl <ADDRESS>/hello
      ```
     预期输出：
      ```bash
      {"hello":"coffee"}
      ```

### 配置健康检查

ALB Ingress支持配置健康检查，可以通过设置以下注解实现。

配置健康检查的YAML示例如下所示：

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: cafe-ingress
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/address-type: internet
    alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"
    alb.ingress.kubernetes.io/healthcheck-enabled: "true"
    alb.ingress.kubernetes.io/healthcheck-path: "/"
    alb.ingress.kubernetes.io/healthcheck-protocol: "HTTP"
    alb.ingress.kubernetes.io/healthcheck-method: "HEAD"
    alb.ingress.kubernetes.io/healthcheck-httpcode: "http_2xx"
    alb.ingress.kubernetes.io/healthcheck-timeout-seconds: "5"
    alb.ingress.kubernetes.io/healthcheck-interval-seconds: "2"
    alb.ingress.kubernetes.io/healthy-threshold-count: "3"
    alb.ingress.kubernetes.io/unhealthy-threshold-count: "3"
spec:
  rules:
  - http:
      paths:
      # 配置Context Path。
      - path: /tea
        backend:
          serviceName: tea-svc
          servicePort: 80
      # 配置Context Path。
      - path: /coffee
        backend:
          serviceName: coffee-svc
          servicePort: 80
```

相关参数解释如下表所示。
<table class="table"  >
   <thead class="thead" >
      <tr >
         <th class="entry">参数</th>
         <th class="entry">说明</th>
      </tr>
   </thead>
   <tbody class="tbody">
      <tr >
         <td class="entry"><span style='font-weight:700' >alb.ingress.kubernetes.io/healthcheck-enabled</span></td>
         <td class="entry">（可选）表示是否开启健康检查。默认开启（<span style='font-weight:700' >true</span>）。
         </td>
      </tr>
      <tr >
         <td class="entry" ><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-path</span></td>
         <td class="entry" >（可选）表示健康检查路径。默认<span class="ph filepath" >/</span>。
            <ul class="ul" >
               <li class="li">输入健康检查页面的URL，建议对静态页面进行检查。长度限制为1~80个字符，支持使用字母、数字和短划线（-）、正斜线（/）、半角句号（.）、百分号（%）、半角问号（?）、井号（#）和and（&amp;）以及扩展字符集_;~!（)*[]@$^:',+。URL必须以正斜线（/）开头。</li>
               <li class="li" >HTTP健康检查默认由负载均衡系统通过后端ECS内网IP地址向该服务器应用配置的默认首页发起HTTP Head请求。如果您用来进行健康检查的页面并不是应用服务器的默认首页，需要指定具体的检查路径。</li>
            </ul>
         </td>
      </tr>
      <tr >
         <td class="entry" ><span style='font-weight:700' >alb.ingress.kubernetes.io/healthcheck-protocol</span></td>
         <td class="entry" >（可选）表示健康检查协议。
            <ul class="ul" >
               <li class="li"><span style='font-weight:700' >HTTP</span>（默认）：通过发送HEAD或GET请求模拟浏览器的访问行为来检查服务器应用是否健康。
               </li>
               <li class="li" ><span style='font-weight:700' >TCP</span>：通过发送SYN握手报文来检测服务器端口是否存活。
               </li>
               <li class="li" ><span style='font-weight:700' >GRPC</span>：通过发送POST或GET请求来检查服务器应用是否健康。
               </li>
            </ul>
         </td>
      </tr>
      <tr >
         <td class="entry" ><span style='font-weight:700' >alb.ingress.kubernetes.io/healthcheck-method</span></td>
         <td class="entry" >（可选）选择一种健康检查方法。
            <ul class="ul" >
               <li class="li" ><span style='font-weight:700' >HEAD</span>（默认）：HTTP监听健康检查默认采用HEAD方法。请确保您的后端服务器支持HEAD请求。如果您的后端应用服务器不支持HEAD方法或HEAD方法被禁用，则可能会出现健康检查失败，此时可以使用GET方法来进行健康检查。
               </li>
               <li class="li"><span style='font-weight:700' >POST</span>：GRPC监听健康检查默认采用POST方法。请确保您的后端服务器支持POST请求。如果您的后端应用服务器不支持POST方法或POST方法被禁用，则可能会出现健康检查失败，此时可以使用GET方法来进行健康检查。
               </li>
               <li class="li" ><span style='font-weight:700' >GET</span>：如果响应报文长度超过8 KB，会被截断，但不会影响健康检查结果的判定。
               </li>
            </ul>
         </td>
      </tr>
      <tr >
         <td class="entry" ><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-httpcode</span></td>
         <td class="entry" >设置健康检查正常的状态码。
            <ul class="ul" >
               <li class="li">当健康检查协议为<span style='font-weight:700' >HTTP</span>协议时，可以选择<span style='font-weight:700' >http_2xx</span>（默认）、<span style='font-weight:700' >http_3xx</span>、<span style='font-weight:700' >http_4xx</span>和<span style='font-weight:700' >http_5xx</span>。
               </li>
               <li class="li" >当健康检查协议为<span style='font-weight:700'>GRPC</span>协议时，状态码范围为0~99。支持范围输入，最多支持20个范围值，多个范围值使用半角逗号（,）隔开。
               </li>
            </ul>
         </td>
      </tr>
      <tr >
         <td class="entry"><span style='font-weight:700' >alb.ingress.kubernetes.io/healthcheck-timeout-seconds</span></td>
         <td class="entry" >表示接收健康检查的响应需要等待的时间。如果后端ECS在指定的时间内没有正确响应，则判定为健康检查失败。时间范围为1~300秒，默认值为5秒。</td>
      </tr>
      <tr >
         <td class="entry" ><span style='font-weight:700' >alb.ingress.kubernetes.io/healthcheck-interval-seconds</span></td>
         <td class="entry" >健康检查的时间间隔。取值范围1~50秒，默认为2秒。</td>
      </tr>
      <tr >
         <td class="entry" ><span style='font-weight:700' >alb.ingress.kubernetes.io/healthy-threshold-count</span></td>
         <td class="entry" >表示健康检查连续成功所设置的次数后会将后端服务器的健康检查状态由失败判定为成功。取值范围2~10，默认为3次。</td>
      </tr>
      <tr >
         <td class="entry" ><span style='font-weight:700'>alb.ingress.kubernetes.io/unhealthy-threshold-count</span></td>
         <td class="entry" >表示健康检查连续失败所设置的次数后会将后端服务器的健康检查状态由成功判定为失败。取值范围2~10，默认为3次。</td>
      </tr>
   </tbody>
</table>

### 配置自动发现HTTPS证书功能

ALB Ingress Controller提供证书自动发现功能。您需要首先在SSL证书控制台创建证书，然后ALB Ingress Controller会根据Ingress中TLS配置的域名自动匹配发现证书。

1. 执行以下命令，通过openssl创建证书。
    ```bash
    openssl genrsa -out albtop-key.pem 4096
    openssl req -subj "/CN=demo.alb.ingress.top" -sha256  -new -key albtop-key.pem -out albtop.csr
    echo subjectAltName = DNS:demo.alb.ingress.top > extfile.cnf
    openssl x509 -req -days 3650 -sha256 -in albtop.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out albtop-cert.pem -extfile extfile.cnf
    ```
2. 在SSL证书控制台上传证书。具体操作，请参见上传证书。
3. 在Ingress的YAML中添加以下命令，配置该证书对应的域名。
    ```yaml
    tls:
    - hosts:
        - demo.alb.ingress.top
    ```
   示例如下：
   ```yaml
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: demo-service-https
     namespace: default
   spec:
     ports:
       - name: port1
         port: 443
         protocol: TCP
         targetPort: 8080
     selector:
       app: demo-cafe
     sessionAffinity: None
     type: NodePort
   
   ---
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: demo-cafe
     namespace: default
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: demo-cafe
     template:
       metadata:
         labels:
           app: demo-cafe
       spec:
         containers:
           - image: registry.cn-hangzhou.aliyuncs.com/alb-sample/cafe:v1
             imagePullPolicy: IfNotPresent
             name: demo-cafe
             ports:
               - containerPort: 8080
                 protocol: TCP
   ---
   apiVersion: networking.k8s.io/v1beta1
   kind: Ingress
   metadata:
     annotations:
       alb.ingress.kubernetes.io/address-type: internet
       alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"
       kubernetes.io/ingress.class: alb
     name: demo-https
     namespace: default
   spec:
     #配置证书对应的域名。
     tls:
     - hosts:
       - demo.alb.ingress.top
     rules:
       - host: demo.alb.ingress.top
         http:
           paths:
             - backend:
                 serviceName: demo-service-https
                 servicePort: 443
               path: /
               pathType: Prefix
   ```
4. 执行以下命令，查看证书。
    ```bash
    curl https://demo.alb.ingress.top/tea
    ```
   预期输出：
    ```bash
    {"hello":"tee"}
    ```

### 配置HTTP重定向至HTTPS

ALB Ingress通过设置注解alb.ingress.kubernetes.io/ssl-redirect: "true"，可以将HTTP请求重定向到HTTPS 443端口。

配置示例如下：

```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: demo-service-ssl
  namespace: default
spec:
  ports:
    - name: port1
      port: 80
      protocol: TCP
      targetPort: 8080
  selector:
    app: demo-ssl
  sessionAffinity: None
  type: NodePort
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-ssl
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: demo-ssl
  template:
    metadata:
      labels:
        app: demo-ssl
    spec:
      containers:
        - image: registry.cn-hangzhou.aliyuncs.com/alb-sample/cafe:v1
          imagePullPolicy: IfNotPresent
          name: demo-ssl
          ports:
            - containerPort: 8080
              protocol: TCP
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"
    alb.ingress.kubernetes.io/ssl-redirect: "true"
  name: demo-ssl
  namespace: default
spec:
  tls:
  - hosts:
    - ssl.alb.ingress.top
  rules:
    - host: ssl.alb.ingress.top
      http:
        paths:
          - backend:
              serviceName: demo-service-ssl
              servicePort: 80
            path: /
            pathType: Prefix
```

### 通过注解实现灰度发布

ALB提供复杂路由处理能力，支持基于Header、Cookie以及权重的灰度发布功能。灰度发布功能可以通过设置注解来实现，为了启用灰度发布功能，需要设置注解alb.ingress.kubernetes.io/canary: "true"，通过不同注解可以实现不同的灰度发布功能。
```
说明: 灰度优先级顺序：基于Header>基于Cookie>基于权重（从高到低）。
```
<table class="table">
   <thead class="thead">
      <tr>
         <th class="entry">参数</th>
         <th class="entry">说明</th>
         <th class="entry">示例</th>
      </tr>
   </thead>
   <tbody class="tbody">
      <tr>
         <td class="entry"><code >alb.ingress.kubernetes.io/canary-by-header</code>和<code >alb.ingress.kubernetes.io/canary-by-header-value</code></td>
         <td class="entry">匹配的Request Header的值，该规则允许您自定义Request Header的值，但必须与<code >alb.ingress.kubernetes.io/canary-by-header</code>一起使用。
            <ul class="ul">
               <li class="li">当请求中的<code >header</code>和<code >header-value</code>与设置的值匹配时，请求流量会被分配到灰度服务入口。
               </li>
               <li class="li">对于其他<code >header</code>值，将会忽略<code >header</code>，并通过灰度优先级将请求流量分配到其他规则设置的灰度服务。
               </li>
            </ul>
         </td>
         <td class="entry">当请求Header为<code >location: hz</code>时将访问灰度服务；其它Header将根据灰度权重将流量分配给灰度服务。
    <div class="code-block">
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
      <div class="code-tools">
        <i class="theme-switch-btn"></i><i class="copy-btn"></i>
      </div>
      <pre class="pre codeblock"><code class="hljs language-bash">apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/albconfig.order: <span class="hljs-string">"1"</span>
    alb.ingress.kubernetes.io/vswitch-ids: <span class="hljs-string">"vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"</span>
    alb.ingress.kubernetes.io/canary: <span class="hljs-string">"true"</span>
    alb.ingress.kubernetes.io/canary-by-header: <span class="hljs-string">"location"</span>
    alb.ingress.kubernetes.io/canary-by-header-value: <span class="hljs-string">"hz"</span></code></pre>
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;margin-bottom: 16px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
    </div>
    </td>
   </tr>
   <tr>
      <td class="entry"><code >alb.ingress.kubernetes.io/canary-by-cookie</code></td>
      <td class="entry">基于Cookie的流量切分：
         <ul class="ul">
            <li class="li">当配置的<code >cookie</code>值为<code >always</code>时，请求流量将被分配到灰度服务入口。
            </li>
            <li class="li">当配置的<code >cookie</code>值为<code >never</code>时，请求流量将不会分配到灰度服务入口。
            </li>
         </ul>
         <div class="note note note-note">
            <div class="note-icon-wrapper"><i class="icon-note note"></i></div>
            <div class="note-content"><strong>说明</strong> 基于Cookie的灰度不支持设置自定义，只有<code >always</code>和<code >never</code>。
            </div>
         </div>
      </td>
      <td class="entry">请求的Cookie为<code >demo=always</code>时将访问灰度服务。
    <div class="code-block">
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
      <div class="code-tools">
        <i class="theme-switch-btn"></i><i class="copy-btn"></i>
      </div>
      <pre class="pre codeblock"><code class="hljs language-bash">apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/albconfig.order: <span class="hljs-string">"2"</span>
    alb.ingress.kubernetes.io/address-type: internet
    alb.ingress.kubernetes.io/vswitch-ids: <span class="hljs-string">"vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"</span>
    alb.ingress.kubernetes.io/canary: <span class="hljs-string">"true"</span>
    alb.ingress.kubernetes.io/canary-by-cookie: <span class="hljs-string">"demo"</span></code></pre>
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;margin-bottom: 16px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
    </div>
    </td>
   </tr>
   <tr>
      <td class="entry"><code >alb.ingress.kubernetes.io/canary-weight</code></td>
      <td class="entry">设置请求到指定服务的百分比（值为0~100的整数）。</td>
      <td class="entry">配置灰度服务的权重为50%。
    <div class="code-block">
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
      <div class="code-tools">
        <i class="theme-switch-btn"></i><i class="copy-btn"></i>
      </div>
      <pre class="pre codeblock"><code class="hljs language-bash">apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/albconfig.order: <span class="hljs-string">"3"</span>
    alb.ingress.kubernetes.io/address-type: internet
    alb.ingress.kubernetes.io/vswitch-ids: <span class="hljs-string">"vsw-2zeqgkyib34gw1fxs****,vsw-2zefv5qwao4przzlo****"</span>
    alb.ingress.kubernetes.io/canary: <span class="hljs-string">"true"</span>
    alb.ingress.kubernetes.io/canary-weight: <span class="hljs-string">"50"</span></code></pre>
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;margin-bottom: 16px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
    </div>
    </td>
      </tr>
   </tbody>
</table>



