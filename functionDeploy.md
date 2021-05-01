### How do I deploy a deep learning function?
Each nuclio function must have an entrypoint which should be described in the .yaml file. The yaml file also includes commands for downloading potential dependencies. See a full description [here](https://nuclio.io/docs/latest/tasks/deploying-functions/).
1. Create a folder containing your python/go/java function and a yaml file. E.x. `main.py` and `function.yaml` as seen in [our example](https://github.com/Frans-Lukas/DLScheduler/tree/main/ResNet50).
2. The python file must contain a handler function with two arguments, `context`, and `event`. The handler function name is then given in the yaml file at the `handler` tag.  
3. Add metadata and spec to your yaml file. E.x:
```
apiVersion: "nuclio.io/v1"
kind: "NuclioFunction"
metadata:
  name: res-net-50
  namespace: nuclio
spec:
  runtime: python
  handler: main:train_epoch
  build:
    commands:
    - "mkdir -p /tmp/resnetmodel"
    - "pip install dload pillow keras numpy tensorflow"
```
3. The yaml file should contain all the metadata `nuctl` needs to deploy the funciton. Simply call:
    1. `nuctl deploy --path /path/to/your/function/folder --registry $(minikube ip):5000 --run-registry localhost:5000 -n nuclio` to deploy locally or
    2. `nuctl deploy --path /path/to/your/function/folder	--registry docker.io/$USERNAME -n nuclio` to deploy on dockerhub. 
    3. Sudo might be required for deployment, if sudo is used, it must be used when listing and invoking the funciton. 

### Invoke a nuctl function
1. First list the available functions with `nuctl get function -n nuclio`.
2. Make sure the function has a node port. If it does, invoke the function with `nuctl invoke my-function -n nuclio`.
3. If it does not, see the nuclio installation instructions on how to provide a node port.  

### Debug a nuctl function
Sometimes it is necessary to enter the pod in which the function is deployed for debugging purposes. To create a usable shell inside a pod simply call:
1. List pods to get $POD_NAME with `kubectl get pod -n nuclio`.
2. `kubectl exec --stdin --tty $POD_NAME -n nuclio -- /bin/bash`
