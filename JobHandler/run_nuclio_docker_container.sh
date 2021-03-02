#!/bin/bash

# make sure rbac roles have been applied (see https://nuclio.io/docs/latest/setup/k8s/getting-started-k8s/)
# kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio-rbac.yaml
sudo nuctl deploy $1 --run-image $2 --file function.yaml -n nuclio
