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
6. Enable metrics server `minikube addons enable metrics-server`, the metrics server can take some time to start, which is why we do it this early
    1. Check if metrics server has started and is running with `kubectl -n kube-system get deployments.apps`, if it has continue with these steps otherwise come back here later.
    2. Increase how frequently the metric server scrapes for metrics (by default it seems to only be every 60 seconds).
        Edit the deployment details for metric server `kubectl -n kube-system edit deployments.apps metrics-server`
        Add `- --metric-resolution=15s` where it looks similar to:
        ```
        spec:
            containers:
            - command:
              - /metrics-server
              - --metric-resolution=15s
              - --source=kubernetes.summary_api:https://kubernetes.default?kubeletHttps=true&kubeletPort=10250&insecure=true
              image: k8s.gcr.io/metrics-server-amd64:v0.2.1
              imagePullPolicy: Always
              name: metrics-server
        ```
    3. Check that is running again `kubectl get pods -n kube-system | grep metrics-server`
    4. `kubectl top pods -n nuclio` should now update about every 15 seconds ([faster updates are not recommended, as this is the resolution of metrics calculated within Kubelet](https://github.com/kubernetes-sigs/metrics-server/blob/master/FAQ.md))
7. Create the nuclio namespace with `kubectl create namespace nuclio`
8. Kubernetes with docker requires container images to be stored in a repository, either:
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
9. Create RBAC roles required for Nuclio, `kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio-rbac.yaml`
10. Deploy nuclio to the cluster `kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio.yaml`
11. Use the command `kubectl get pods --namespace nuclio` to verify both the controller and dashboard are running.
12. Forward the Nuclio dashboard port `kubectl port-forward -n nuclio $(kubectl get pods -n nuclio -l nuclio.io/app=dashboard -o jsonpath='{.items[0].metadata.name}') 8070:8070`
13. Browse to http://localhost:8070 or
14. Deploy a function with the Nuclio CLI (nuctl) (for local repositories thr URL is `--registry $(minikube ip):5000 --run-registry localhost:5000`, if we use docker hub the URL is `--registry docker.io/<username>` with your dockerhub username) `nuctl deploy helloworld -n nuclio -p https://raw.githubusercontent.com/nuclio/nuclio/master/hack/examples/golang/helloworld/helloworld.go --http-trigger-service-type=NodePort <URL>`
15. Get function info `nuctl get function helloworld` if NodePort is zero, go to 16, otherwise:
16. Invoke function with `nuctl invoke helloworld --method POST --body '{"hello":"world"}' --content-type "application/json"`
17. (If nodeport is zero), `kubectl patch svc nuclio-helloworld -p '{"spec": {"type": "NodePort"}}' -n nuclio`
18. (If nodeport is zero), Get the current nodeport of the service with `kubectl get svc -n nuclio` (Column PORT(s), with format PORT:NODEPORT/TCP)
19. (If nodeport is zero), invoke function with `curl $(minikube ip):PORT`
