# cloud-provider-alibaba-cloud

## Publishing gcp-controller-manager image

This command will build and publish
`gcr.io/k8s-image-staging/gcp-controller-manager:latest`:

```
bazel run //cmd/gcp-controller-manager:publish
```

Environment variables `IMAGE_REPO` and `IMAGE_TAG` can be used to override
destination GCR repository and tag.

This command will build and publish
`gcr.io/my-repo/gcp-controller-manager:v1`:


```
IMAGE_REPO=my-repo IMAGE_TAG=v1 bazel run //cmd/gcp-controller-manager:publish
```
