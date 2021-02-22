#!/bin/bash

# make sure rbac roles have been applied (see https://nuclio.io/docs/latest/setup/k8s/getting-started-k8s/)
# kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio-rbac.yaml
./run_nuclio_docker_container.sh lenet2-1
./run_nuclio_docker_container.sh lenet2-2
./run_nuclio_docker_container.sh lenet2-3
./run_nuclio_docker_container.sh real_para_server2