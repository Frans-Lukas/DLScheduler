#!/bin/bash

if [ $# -eq 0 ]; then
  echo "No arguments supplied"
fi

# make sure rbac roles have been applied (see https://nuclio.io/docs/latest/setup/k8s/getting-started-k8s/)
# kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio-rbac.yaml
sudo nuctl deploy lenet --run-image docker.io/franslukas/processor-lenet:latest --file ./worker/function.yaml -n nuclio
