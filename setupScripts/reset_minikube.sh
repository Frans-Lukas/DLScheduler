#!/bin/bash


minikube delete
minikube start
kubectl create namespace nuclio
kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio.yaml
sleep 5
kubectl get pods --namespace nuclio

nuctl deploy helloworld -n nuclio -p https://raw.githubusercontent.com/nuclio/nuclio/master/hack/examples/golang/helloworld/helloworld.go --http-trigger-service-type=NodePort --registry docker.io/franslukas