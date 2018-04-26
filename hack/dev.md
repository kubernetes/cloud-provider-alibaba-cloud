## Dev

For test

```
sudo ssh -v -N -L 127.0.0.1:80:127.0.0.1:30977 root@106.15.161.115

```

```
export ACCESS_KEY_ID
export ACCESS_KEY_SECRET
export METADATA_ENDPOINT

--kubeconfig=/etc/kubernetes/cloud-controller-manager.conf
--address=127.0.0.1
--allow-untagged-cloud=true
--leader-elect=true
--cloud-provider=alicloud
--allocate-node-cidrs=true
--cluster-cidr=172.16.0.0/16
--use-service-account-credentials=true
--route-reconciliation-period=30s
--v=5
```

```
/var/run/secrets/kubernetes.io/serviceaccount/token
```
