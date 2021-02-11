### How do i deploy a new scheduler?
This tutorial will allow you to deploy another scheduler alongside the default kubernetes scheduler.
The new scheduler deployed here will work the same way as the default one.
How to make it behave differently will be looked into later.

Follow the [official tutorial](https://kubernetes.io/docs/tasks/extend-kubernetes/configure-multiple-schedulers/).
But if you want to use dockerhub instead of Google Container Registry, make the modificatons described below.
1. When building and pushing to the registry, use these commands instead:
```
docker build -t <user>/<registry>:<tag> <path to Dockerfile>
docker push <user>/<registry>:<tag>
```
2. Modify `my-scheduler.yaml`, specifically the line `image: gcr.io/my-gcp-project/my-kube-scheduler:1.0` should be changed to:
`image: <user>/<registry>:<tag>`
