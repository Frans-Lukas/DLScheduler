#!/bin/bash

# p03.ds.cs.umu.se
# p04.ds.cs.umu.se
# p05.ds.cs.umu.se

sudo modprobe br_netfilter
sudo apt update
# Kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release

sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update

# kubeadm, kubectl, kubelet
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
sudo apt-get update
# docker
sudo apt-get install -y docker-ce docker-ce-cli containerd.io
sudo usermod -aG docker pirat
sudo service docker start
sudo swapoff -a

#nuclio
curl -s https://api.github.com/repos/nuclio/nuclio/releases/latest | grep -i "browser_download_url.*nuctl.*$(uname)" | cut -d : -f 2,3 | tr -d \" | wget -O nuctl -qi - && chmod +x nuctl


kubectl create namespace nuclio
sudo docker login
kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/nuclio/nuclio/master/hack/k8s/resources/nuclio.yaml
kubectl get pods --namespace nuclio


# sudo docker run hello-world

# master node setup:
#sudo cp /etc/kubernetes/admin.conf $HOME/
#sudo chown $(id -u):$(id -g) $HOME/admin.conf
#export KUBECONFIG=$HOME/admin.conf

# to run main:
# go run main.go singleTenant83.json test.txt /etc/kubernetes/admin.conf


# Weavenet CNI plugin
# https://www.weave.works/docs/net/latest/kubernetes/kube-addon/
# kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')"

# all other nodes:
#sudo kubeadm join 130.239.48.222:6443 --token v64mwj.9pqsiwixsup2qq5n --discovery-token-ca-cert-hash sha256:68c0bc5173dd31c78f8e2a939cf740d3746e0220d8b2b139d3be21617df215c9
