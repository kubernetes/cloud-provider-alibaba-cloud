# Alibaba Cloud Provider

# Application Load Balancer Ingress introduction

Application Load Balancer (ALB) Ingresses are compatible with NGINX Ingresses, and provide improved traffic routing capabilities based on ALB instances. ALB Ingresses support complex routing, automatic certificate discovery, and the HTTP, HTTPS, and Quick UDP Internet Connection (QUIC) protocols. ALB Ingresses meet the requirements of cloud-native applications for ultra-high elasticity and balancing of heavy traffic loads at Layer 7. This topic describes how to use an ALB Ingress to expose Services.

## Background information

An Ingress provides a collection of rules that manage external access to Services in a cluster. You can configure forwarding rules to assign Services different externally-accessible URLs. However, NGINX Ingresses and Layer 4 Server Load Balancer (SLB) Ingresses cannot meet the requirements of cloud-native applications, such as complex routing, support for multiple application layer protocols (such as QUIC), and balancing of heavy traffic loads at Layer 7.

## Configure Albconfig objects

An Albconfig object is a CustomResourceDefinition (CRD) object that Container Service for Kubernetes (ACK) provides for the Application Load Balancer (ALB) Ingress controller. The ALB Ingress controller uses Albconfig objects to configure ALB instances and listeners. This topic describes how to create and modify an Albconfig object, and how to specify an Albconfig object for an Ingress.

### Background information
The ALB Ingress controller retrieves the changes to Ingresses from the API server and dynamically generates Albconfig objects when Ingresses changes are detected. Then, the ALB Ingress controller performs the following operations in sequence: create ALB instances, configure listeners, create Ingress rules, and configure backend server groups. The Service, Ingress, and Albconfig objects interact with each other in the following ways:
- A Service is an abstraction of an application that is deployed on a set of replicated pods.
- An Ingress contains reverse proxy rules. It controls to which Services HTTP or HTTPS requests are routed. For example, an Ingress routes requests to different Services based on the hosts and URLs in the requests.
- An Albconfig is a CustomResourceDefinition (CRD) object that the ALB Ingress controller uses to configure ALB instances and listeners. An Albconfig corresponds to one ALB instance.
- An Albconfig object is used to configure an ALB instance. The ALB instance can be specified in forwarding rules of multiple Ingresses. Therefore, an Albconfig object can be associated with multiple Ingresses.

### Create an Albconfig object

When you create an Ingress, an Albconfig object named default is automatically created in the kube-system namespace.

1. Use the following template to create an Ingress and an Albconfig object:
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
         # Configure a context path.
         - path: /tea
           backend:
             serviceName: tea-svc
             servicePort: 80
         # Configure a context path.
         - path: /coffee
           backend:
             serviceName: coffee-svc
             servicePort: 80
   ```
2. Run the following command to query the automatically created Albconfig object:
    ```bash
    kubectl -n kube-system get albconfig
    ```
   output：
    ```bash
    NAME      AGE
    default   87m
    ```
   The following content shows the configurations of the default Albconfig object:
   ```yaml
   apiVersion: alibabacloud.com/v1
   kind: AlbConfig
   metadata:
     name: default                      # The name of the Albconfig object. 
     namespace: kube-system             # The namespace to which the Albconfig object belongs. 
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
       - vSwitchId: vsw-wz92lvykqj1siwvif****        # A vSwitch that is specified for the Albconfig object. You must specify two vSwitches. 
       - vSwitchId: vsw-wz9mnucx78c7i6iog****        # A vSwitch that is specified for the Albconfig object. 
   status:
     loadBalancer:
       dnsname: alb-s2em8fr9debkg5****.cn-shenzhen.alb.aliyuncs.com
       id: alb-s2em8fr9debkg5****
   ```
### Change the name of an Albconfig object
To change the name of an Albconfig object, run the following command. The change is automatically applied after you save the modification.
```bash
kubectl -n kube-system edit albconfig default
...
  spec:
    config:
      name: basic   # The new name that you want to use. 
...
```
### Change the vSwitches that are specified for an Albconfig object
To change the vSwitches that are specified for an Albconfig object, run the following command. The change is automatically applied after you save the modification.
```bash
kubectl -n kube-system edit albconfig default
...
  zoneMappings:
    - vSwitchId: vsw-wz92lvykqj1siwvif****
    - vSwitchId: vsw-wz9mnucx78c7i6iog****
...
```
### Specify an Albconfig object for an Ingress
To specify an Albconfig object for an Ingress, use the annotation alb.ingress.kubernetes.io/albconfig.name. This allows you to use a specific ALB instance.
```
Note If the specified Albconfig object does not exist in the kube-system namespace, the system automatically creates an Albconfig object named default in the kube-system namespace.
```
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: cafe-ingress-v1
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/address-type: internet
    alb.ingress.kubernetes.io/albconfig.name: basic
    alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"
spec:
  rules:
  - http:
      paths:
      # Configure a context path.
      - path: /tea
        backend:
          serviceName: tea-svc
          servicePort: 80
      # Configure a context path.
      - path: /coffee
        backend:
          serviceName: coffee-svc
          servicePort: 80
```
### Delete an ALB instance
An Albconfig object is used to configure an ALB instance. Therefore, you can delete an ALB instance by deleting the corresponding Albconfig object. Before you can delete an Albconfig object, you must delete all Ingresses that are associated with the Albconfig object.
```bash
kubectl -n kube-system delete albconfig default
```
Replace default with the name of the Albconfig object that you want to delete.

## Expose Services by using an ALB Ingress

This topic describes how to use an ALB Ingress to expose Services.

### Background information

An Ingress provides a collection of rules that manage external access to Services in a cluster. You can configure forwarding rules to assign Services different externally-accessible URLs. However, NGINX Ingresses and Layer 4 Server Load Balancer (SLB) Ingresses cannot meet the requirements of cloud-native applications, such as complex routing, support for multiple application layer protocols (such as QUIC), and balancing of heavy traffic loads at Layer 7.

### Step 1: Deploy Services

1. Create a cafe-service.yaml file and copy the following content to the file. The file is used to deploy two Deployments named coffee and tea and two Services named coffee and tea.
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
     clusterIP: None
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
     clusterIP: None
   ```
2. Run the following command to deploy the Deployments and Services:
    ```bash
    kubectl apply -f cafe-service.yaml
    ```
   Expected output:
    ```bash
    deployment "coffee" created
    service "coffee-svc" created
    deployment "tea" created
    service "tea-svc" created
    ```
3. Run the following command to query the status of the Services:
    ```bash
    kubectl get svc,deploy
    ```
   Expected output:
    ```bash
    NAME             TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
    coffee-svc   NodePort    172.16.231.169   <none>        80:31124/TCP   6s
    tea-svc      NodePort    172.16.38.182    <none>        80:32174/TCP   5s
    NAME            DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
    deploy/coffee   2         2         2            2           1m
    deploy/tea      1         1         1            1           1m
    ```

Step 2: Configure an Ingress
1. Create a cafe-ingress.yaml and copy the following content to the file:
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
         #Configure context path.
         - path: /tea
           backend:
             serviceName: tea-svc
             servicePort: 80
         #Configure context path.
         - path: /coffee
           backend:
             serviceName: coffee-svc
             servicePort: 80
   ```
   <table class="table">
   <thead class="thead">
      <tr>
         <th class="entry">Parameter</th>
         <th class="entry">Description</th>
      </tr>
   </thead>
   <tbody class="tbody">
      <tr>
         <td class="entry">(Optional) <span style='font-weight:700'>alb.ingress.kubernetes.io/name</span></td>
         <td class="entry">The name of the ALB instance that you want to use. </td>
      </tr>
      <tr>
         <td class="entry" >(Optional) <span style='font-weight:700'>alb.ingress.kubernetes.io/address-type</span></td>
         <td class="entry">The type of IP address that the ALB instance uses to provide services. Valid values:
            <ul class="ul">
               <li class="li">Internet: The ALB instance uses a public IP address. The domain name of the Ingress is resolved to the public IP address of the ALB instance. Therefore, the ALB instance is accessible over the Internet. This is the default value. 
               </li>
               <li class="li">Intranet: The ALB instance uses a private IP address. The domain name of the Ingress is resolved to the private IP address of the ALB instance. Therefore, the ALB instance is accessible only within the virtual private cloud (VPC) where the ALB instance is deployed. 
               </li>
            </ul>
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/vswitch-ids</span></td>
         <td class="entry">The IDs of the vSwitches that are used by the ALB Ingress. You must specify at least two vSwitch IDs and the vSwitches must be deployed in different zones. For more information about the regions and zones that are supported by ALB Ingresses, see <span><a title="This topic describes the regions and zones that are supported by ." href="https://www.alibabacloud.com/help/zh/doc-detail/258300.htm#task-2087008">Supported regions and zones</a></span>. 
         </td>
      </tr>
   </tbody>
   </table>

2. Run the following command to configure an externally-accessible domain name and a path for the coffee and tea Services separately:
    ```bash
    kubectl apply -f cafe-ingress.yaml
    ```
   Expected output:
    ```bash
    ingress "cafe-ingress" created
    ```
3. Run the following command to query the IP address of the ALB instance:
    ```bash
    kubectl get ing
    ```
   Expected output:
    ```
    NAME           CLASS    HOSTS   ADDRESS                                               PORTS   AGE
    cafe-ingress   <none>   *       alb-m551oo2zn63yov****.cn-hangzhou.alb.aliyuncs.com   80      50s
    ```

### Step 3: Access the Services
- After you obtain the IP address of the ALB instance, Access the coffee Service by using a CLI.
    ```bash
    curl http://alb-m551oo2zn63yov****.cn-hangzhou.alb.aliyuncs.com/coffee
    ```
- After you obtain the IP address of the ALB instance, Access the tea Service by using a CLI.
    ```bash
    curl http://alb-m551oo2zn63yov****.cn-hangzhou.alb.aliyuncs.com/tea
    ```
## Advanced ALB Ingress configurations

An Ingress is an API object that you can use to provide Layer 7 load balancing to manage external access to Services in a serverless Kubernetes (ASK) cluster. This topic describes how to use Application Load Balancer (ALB) Ingresses to forward requests to backend server groups based on domain names and URL paths, redirect HTTP requests to HTTPS, and implement canary releases.

### Forward requests based on domain names
Perform the following steps to create a simple Ingress with or without a domain name to forward requests.
- Create a simple Ingress with a domain name.
    1. Use the following template to create a Deployment, a Service, and an Ingress. Requests to the domain name of the Ingress are forwarded to the Service.
   ```
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
     type: ClusterIP
   
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
       - host: demo.domain.ingress.top  # The domain name of the Ingress. 
         http:
           paths:
             - backend:
                 serviceName: demo-service
                 servicePort: 80
               path: /hello
               pathType: ImplementationSpecific
   ```
    2. Run the following command to access the application by using the specified domain name.

       Replace ADDRESS with the address of the related ALB instance. You can obtain the address by running the kubectl get ing command.
        ```bash
        curl -H "host: demo.domain.ingress.top" <ADDRESS>/hello
        ```
       Expected output:
        ```bash
        {"hello":"coffee"}
        ```

- Create a simple Ingress without a domain name.
    1. Use the following template to create an Ingress:
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
    2. Run the following command to access the application without using a domain name.

       Replace ADDRESS with the address of the related ALB instance. You can obtain the address by running the kubectl get ing command.
        ```bash
        curl <ADDRESS>/hello
        ```
       Expected output:
        ```bash
        {"hello":"coffee"}
        ```

### Forward requests based on URL paths

ALB Ingresses support request forwarding based on URL paths. You can use the pathType parameter to configure different URL match policies. The valid values of pathType are Exact, ImplementationSpecific, and Prefix.

The following steps show how to configure different URL match policies.
- Exact：exactly matches the URL path with case sensitivity.

    1. Use the following template to create an Ingress:
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
    2. Run the following command to access the application.

       Replace ADDRESS with the address of the related ALB instance. You can obtain the address by running the kubectl get ing command.
        ```bash
        curl <ADDRESS>/hello
        ```
       Expected output:
        ```bash
        {"hello":"coffee"}
        ```
- ImplementationSpecific: the default match policy. For ALB Ingresses, the ImplementationSpecific policy has the same effect as the Exact policy. However, the controllers of Ingresses with the ImplementationSpecific policy and the controllers Ingresses with the Exact policy are implemented in different ways.

    1. Use the following template to create an Ingress:
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
    2. Run the following command to access the application.

       Replace ADDRESS with the address of the related ALB instance. You can obtain the address by running the kubectl get ing command.
        ```bash
        curl <ADDRESS>/hello
        ```
       Expected output:
        ```bash
        {"hello":"coffee"}
        ```
- Prefix: matches based on a URL path prefix separated by forward slashes (/). The match is case-sensitive and performed on each element of the path.

    1. Use the following template to create an Ingress:
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
    2. Run the following command to access the application.

       Replace ADDRESS with the address of the related ALB instance. You can obtain the address by running the kubectl get ing command.
        ```bash
        curl <ADDRESS>/hello
        ```
       Expected output:
        ```bash
        {"hello":"coffee"}
        ```

### Configure health checks

You can configure health checks for ALB Ingresses by using the following annotations.

The following YAML template provides an example on how to create an Ingress that has health check enabled:
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
      # Configure a context path. 
      - path: /tea
        backend:
          serviceName: tea-svc
          servicePort: 80
      # Configure a context path. 
      - path: /coffee
        backend:
          serviceName: coffee-svc
          servicePort: 80
```
The following table describes the parameters in the YAML template.
<table class="table">
   <thead class="thead">
      <tr>
         <th class="entry">Parameter</th>
         <th class="entry">Description</th>
      </tr>
   </thead>
   <tbody class="tbody">
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-enabled</span></td>
         <td class="entry">Optional. Specifies whether to enable health check. Default value: <span style='font-weight:700'>true</span>. 
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-path</span></td>
         <td class="entry">Optional. The path at which health checks are performed. Default value: <span class="ph filepath">/</span>. 
            <ul class="ul">
               <li class="li" >Enter the URL of the web page on which you want to perform health checks. We recommend that you enter the URL of a static web page. The URL must be 1 to 80 characters in length, and can contain letters, digits, hyphens (-), forward slashes (/), periods (.), percent signs (%), question marks (?), number signs (#), and ampersands (&amp;). The URL can also contain the following special characters: _ ; ~ ! ( ) * [ ] @ $ ^ : ' , +. The URL must start with a forward slash (/). 
               </li>
               <li class="li">By default, to perform health checks, the ALB instance sends HTTP HEAD requests to the default application homepage configured on the backend Elastic Compute Service (ECS) instance. The ALB instance sends the requests to the private IP address of the ECS instance. If you do not want to use the default application homepage for health checks, you must specify a path. 
               </li>
            </ul>
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-protocol</span></td>
         <td class="entry">Optional. The protocol used by health checks. 
            <ul class="ul">
               <li class="li"><span style='font-weight:700'>HTTP</span>: The ALB instance sends HEAD or GET requests to a backend server to simulate access from a browser and check whether the backend server is healthy. This is the default protocol. 
               </li>
               <li class="li"><span style='font-weight:700'>TCP</span>: The ALB instance sends TCP SYN packets to a backend server to check whether the port of the backend server is available to receive requests. 
               </li>
               <li class="li"><span style='font-weight:700'>GRPC</span>: The ALB instance sends POST or GET requests to a backend server to check whether the backend server is healthy. 
               </li>
            </ul>
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-method</span></td>
         <td class="entry">Optional. The request method used by health checks. 
            <ul class="ul">
               <li class="li"><span style='font-weight:700'>HEAD</span>: By default, HTTP health checks send HEAD requests to a backend server. This is the default method. Make sure that your backend server supports HEAD requests. If your backend server does not support HEAD requests or HEAD requests are disabled, health checks may fail. In this case, you can use GET requests to perform health checks.
               </li>
               <li class="li"><span style='font-weight:700'>POST</span>: By default, gRPC health checks use the POST method. Make sure that your backend servers support POST requests. If your backend server does not support POST requests or POST requests are disabled, health checks may fail. In this case, you can use GET requests to perform health checks. 
               </li>
               <li class="li"><span style='font-weight:700'>GET</span>: If the length of a response packet exceeds 8 KB, the response is truncated. However, the health check result is not affected. 
               </li>
            </ul>
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-httpcode</span></td>
         <td class="entry">Specify the status codes that are returned when health check results are normal. 
            <ul class="ul">
               <li class="li">When the health check protocol is set to <span style='font-weight:700'>HTTP</span>, the valid values are <span style='font-weight:700'>http_2xx</span>, <span style='font-weight:700'>http_3xx</span>, <span style='font-weight:700'>http_4xx</span>, and <span style='font-weight:700'>http_5xx</span>. The default value for HTTP health checks is http_2xx. 
               </li>
               <li class="li">When the health check protocol is set to <span style='font-weight:700'>GRPC</span>, valid values are 0 to 99. Value ranges are supported. You can enter at most 20 value ranges. Separate multiple value ranges with commas (,). 
               </li>
            </ul>
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-timeout-seconds</span></td>
         <td class="entry">Specifies the timeout period of a health check. If a backend server does not respond within the specified timeout period, the server fails the health check. Valid values: 1 to 300. Default value: 5. Unit: seconds. 
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/healthcheck-interval-seconds</span></td>
         <td class="entry">The interval at which health checks are performed. Unit: seconds. Valid values: 1 to 50. Default value: 2. Unit: seconds. 
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/healthy-threshold-count</span></td>
         <td class="entry">Specifies the number of times that an unhealthy backend server must consecutively pass health checks before the server is considered healthy. Valid values: 2 to 10. Default value: 3. 
         </td>
      </tr>
      <tr>
         <td class="entry"><span style='font-weight:700'>alb.ingress.kubernetes.io/unhealthy-threshold-count</span></td>
         <td class="entry">Specifies the number of times that a healthy backend server must consecutively fail health checks before the server is considered unhealthy. Valid values: 2 to 10. Default value: 3. 
         </td>
      </tr>
   </tbody>
</table>


### Configure automatic certificate discovery

The ALB Ingress controller supports automatic certificate discovery. You must first create a certificate in the SSL Certificates console. Then, specify the domain name of the certificate in the Transport Layer Security (TLS) configurations of the Ingress. This way, the ALB Ingress controller can automatically match and discover the certificate based on the TLS configurations of the Ingress.

1. Configure automatic certificate discovery
    ```bash
    openssl genrsa -out albtop-key.pem 4096
    openssl req -subj "/CN=demo.alb.ingress.top" -sha256  -new -key albtop-key.pem -out albtop.csr
    echo subjectAltName = DNS:demo.alb.ingress.top > extfile.cnf
    openssl x509 -req -days 3650 -sha256 -in albtop.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out albtop-cert.pem -extfile extfile.cnf
    ```
2. Upload the certificate to the SSL Certificates console. For more information, see Upload certificates.
3. Add the following setting to the YAML template of the Ingress to specify the domain name in the created certificate:
   ```yaml
   tls:
     - hosts:
       - demo.alb.ingress.top
   ```
   The following code block is an example:
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
     type: ClusterIP
   
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
4. Run the following command to query the certificate:
    ```bash
    curl https://demo.alb.ingress.top/tea
    ```
   Expected output:
    ```bash
    {"hello":"tee"}
    ```

### Redirect HTTP requests to HTTPS

To redirect HTTP requests to HTTPS, you can add the alb.ingress.kubernetes.io/ssl-redirect: "true" annotation to the ALB Ingress configurations. This way, HTTP requests are redirected to HTTPS port 443.

The following code block is an example:
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
  type: ClusterIP

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
    alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhasriop,vsw-k1amdv9ax94gr5iwamuwu"
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
### Use annotations to implement canary releases

ALB can handle complex traffic routing scenarios and support canary releases based on request headers, cookies, and weights. You can implement canary releases by adding annotations to Ingress configurations. To enable canary releases, you must add the nginx.ingress.kubernetes.io/canary: "true" annotation. This section describes how to use different annotations to implement canary releases.
```
Note Canary releases that use different rules take effect in the following order: header-based > cookie-based > weight-based.
```
<table class="table">
   <thead class="thead" >
      <tr >
         <th class="entry">Parameter</th>
         <th class="entry">Description</th>
         <th class="entry">Reference</th>
      </tr>
   </thead>
   <tbody class="tbody">
      <tr>
         <td class="entry"><code style='font-family:monospace'>alb.ingress.kubernetes.io/canary-by-header</code> and <code style='font-family:monospace'>alb.ingress.kubernetes.io/canary-by-header-value</code></td>
         <td class="entry">Traffic splitting based on the value of the request header. This parameter must be used in combination with <code style='font-family:monospace'>alb.ingress.kubernetes.io/canary-by-header</code>. 
            <ul class="ul">
               <li class="li">If the <code style='font-family:monospace'>header</code> and <code style='font-family:monospace'>header value</code> of a request match the rule, the request is routed to the new application version.
               </li>
               <li class="li">Requests whose <code style='font-family:monospace'>header</code> values do not match the rule are routed based on rules other than the request header <code style='font-family:monospace'>values</code>. 
               </li>
            </ul>
         </td>
         <td class="entry">If a request has <code style='font-family:monospace'>location: hz</code> specified, the request is routed to the new application version. Otherwise, the request is routed based on other rules other than the request header. 
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
    alb.ingress.kubernetes.io/albconfig.order: "1"
    alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"
    alb.ingress.kubernetes.io/canary: "true"
    alb.ingress.kubernetes.io/canary-by-header: "location"
    alb.ingress.kubernetes.io/canary-by-header-value: "hz"</code></pre>
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;margin-bottom: 16px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
    </div>
    </td>
   </tr>
   <tr>
      <td class="entry"><code style='font-family:monospace'>alb.ingress.kubernetes.io/canary-by-cookie</code></td>
      <td class="entry">Traffic splitting based on cookies:
         <ul class="ul">
            <li class="li">If you set <code style='font-family:monospace'>cookie</code> to <code style='font-family:monospace'>always</code>, requests that match the rule are routed to the new application version. 
            </li>
            <li class="li">If you set <code style='font-family:monospace'>cookie</code> to <code style='font-family:monospace'>never</code>, requests that match the rule are routed to the old application version. 
            </li>
         </ul>
         <br/><i style='background-image:url(//img.alicdn.com/tfs/TB1SrZYn7voK1RjSZFDXXXY3pXa-40-40.png);'></i><strong>Note:</strong>
         <br/>Cookie-based canary release does not support custom settings. The cookie value must be <code style='font-family:monospace'>always</code> or <code style='font-family:monospace'>never</code>. 
      </td>
      <td class="entry">Requests with the <code style='font-family:monospace'>demo=always</code> cookie are routed to the new application version. 
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
    alb.ingress.kubernetes.io/albconfig.order: "2"
    alb.ingress.kubernetes.io/address-type: internet
    alb.ingress.kubernetes.io/vswitch-ids: "vsw-k1akdsmts6njkvhas****,vsw-k1amdv9ax94gr5iwa****"
    alb.ingress.kubernetes.io/canary: "true"
    alb.ingress.kubernetes.io/canary-by-cookie: "demo"</code></pre>
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;margin-bottom: 16px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
    </div>
    </td>
   </tr>
   <tr>
      <td class="entry"><code style='font-family:monospace'>alb.ingress.kubernetes.io/canary-weight</code></td>
      <td class="entry">Traffic splitting based on weights. The weight is a percentage value that ranges from 0 to 100. 
      </td>
      <td class="entry">The following template specifies that 50% of the traffic is routed to the new application version. 
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
    alb.ingress.kubernetes.io/albconfig.order: "3"
    alb.ingress.kubernetes.io/address-type: internet
    alb.ingress.kubernetes.io/vswitch-ids: "vsw-2zeqgkyib34gw1fxs****,vsw-2zefv5qwao4przzlo****"
    alb.ingress.kubernetes.io/canary: "true"
    alb.ingress.kubernetes.io/canary-weight: "50"</code></pre>
      <div class="pre-scrollbar-track" style="display: none;width: 100%;height: 4px;margin-bottom: 16px;">
        <div class="pre-scrollbar-thumb" style="height: 100%;background-color: #d7d8d9;position: relative;"></div>
      </div>
    </div>
    </td>
      </tr>
   </tbody>
</table>