## How to Run E2ETest

Using the following command to start an E2E test
```$bash
 go test -v \
    k8s.io/cloud-provider-alibaba-cloud/cmd/e2e \
    -test.run ^TestE2E$ \
    --kubeconfig /path/to/.kube/config.e2e \
    --cloud-config /path/to/.kube/config.cloud
```

File `config.e2e` is the kubeconfig file used to connect to the target kubernetes cluster.
For Example:
```$yaml
apiVersion: v1
clusters:
- cluster:
    server: https://${APISERVER_ADDR}:6443
    certificate-authority-data: 
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: "kubernetes-admin"
  name: kubernetes-admin-c62df1ad0683947aa9556a09f20493964
current-context: kubernetes-admin-c62df1ad0683947aa9556a09f20493964
kind: Config
preferences: {}
users:
- name: "kubernetes-admin"
  user:
    client-certificate-data: 
    client-key-data: 

```

File `config.cloud` is the cloudprovider cloud-config which mainly used in initializing CloudOpenAPI client.