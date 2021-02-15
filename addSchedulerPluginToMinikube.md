### How do i add a plugin to the minikube scheduler?
Here you will find a tutorial to add a plugin to the scheduler in a minikube cluster

Follow the tutorial [here](https://github.com/kubernetes-sigs/scheduler-plugins/blob/master/doc/install.md).
BUT since we are using minikube some tips for the commands follow here:
1. start the cluster with `minikube start`.
2. enter the master node with `sudo docker exec -it minikube bash`.
3. in `/etc/kubernetes/coscheduling-config.yaml`, `leaderelect:` should be `false`.

### TODO
The above tutorial shows adding the plugin [coscheduling](https://github.com/kubernetes-sigs/scheduler-plugins/tree/master/pkg/coscheduling), figure out how to use our own plugin instead.
