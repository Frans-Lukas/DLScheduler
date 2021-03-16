#!/bin/bash

# make sure rbac roles have been applied (see https://nuclio.io/docs/latest/setup/k8s/getting-started-k8s/)
# kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio-rbac.yaml
nuctl deploy tensorflow --run-image docker.io/franslukas/processor-mxnet:latest --file ./TensorFlow/function.yaml -n nuclio --http-trigger-service-type=NodePort
