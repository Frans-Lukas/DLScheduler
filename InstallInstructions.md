## Install instructions for getting a local environment up and running with minikube and nuclio
1. Install [minikube](https://minikube.sigs.k8s.io/docs/start/), a kubernetes controller
    1. Test with `minikube version`
2. Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/), kubernetes command line interface
    1. Test with `kubectl version`
3. Install [nuctl](https://nuclio.io/docs/latest/reference/nuctl/nuctl/), nuclio command line interface
    1. Test with `nuctl version`
4. Install [docker](https://docs.docker.com/engine/install/), container manager
    1. Test with `docker version`
    2. Add your user to the docker group `sudo usermod -aG docker $USER && newgrp docker`
5. Start minikube `minikube start --driver docker --addons ingress`
6. Create the nuclio namespace with `kubectl create namespace nuclio`
7. Kubernetes with docker requires container images to be stored in a repository, either:
    1. Use local repository with `minikube ssh -- docker run -d -p 5000:5000 registry:2`
    <details>
    <summary>ii. Create a [docker hub](https://hub.docker.com/) and store your credentials with:</summary>
      
  
        read -s mypassword
        enter your password
        kubectl create secret docker-registry registry-credentials --namespace nuclio \
            --docker-username <username> \
            --docker-password $mypassword \
            --docker-server hub.docker.com \
            --docker-email ignored@nuclio.io
        unset mypassword
        
      
    </details>
8. Create RBAC roles required for Nuclio, `kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio-rbac.yaml`
9. Deploy nuclio to the cluster `kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio.yaml`
10. Use the command kubectl get pods --namespace nuclio to verify both the controller and dashboard are running.
11. Forward the Nuclio dashboard port `kubectl port-forward -n nuclio $(kubectl get pods -n nuclio -l nuclio.io/app=dashboard -o jsonpath='{.items[0].metadata.name}') 8070:8070`
12. Browse to http://localhost:8070 or
13. Deploy a function with the Nuclio CLI (nuctl) (Since we use docker hub the url is docker.io/<username> with your dockerhub username) `nuctl deploy helloworld -n nuclio -p https://raw.githubusercontent.com/nuclio/nuclio/master/hack/examples/golang/helloworld/helloworld.go --registry docker.io/<username>`
14. Get function info `nuctl get function helloworld`, if NodePort is zero, go to 16, otherwise:
15. Invoke function with `nuctl invoke helloworld --method POST --body '{"hello":"world"}' --content-type "application/json"`
16. (If nodeport is zero), `kubectl patch svc nuclio-helloworld -p '{"spec": {"type": "NodePort"}}' -n nuclio`
17. (If nodeport is zero), Get the current nodeport of the service with `kubectl get svc -n nuclio` (Column PORT(s), with format PORT:NODEPORT/TCP)
18. (If nodeport is zero), invoke function with `curl $(minikube ip):PORT`
