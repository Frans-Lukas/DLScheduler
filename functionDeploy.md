### How do I deploy a deep learning function?
Each nuclio function must have an entrypoint which should be described in the .yaml file. The yaml file also includes commands for downloading potential dependencies. 
1. Create a folder containing your python/go/java function and a yaml file. E.x. `main.py` and `function.yaml` as seen in [our example](https://github.com/Frans-Lukas/DLScheduler/tree/main/ResNet50).
2. Add metadata and spec to your yaml file. E.x:
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
    1. `nuctl deploy --path /path/to/your/function/folder --registry $(minikube ip):5000 --run-registry localhost:5000` to deploy locally or
    2. `nuctl deploy --path /path/to/your/function/folder	--registry docker.io/$USERNAME` to deploy on dockerhub. 
