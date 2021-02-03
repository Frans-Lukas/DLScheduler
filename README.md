## Multi tenant serverless deep learning scheduler

### Models and Datasets
* [ResNext model](https://github.com/lfz/ResNeXt-DenseNet) on [CIFAR10 dataset](https://www.cs.toronto.edu/~kriz/cifar.html).
* [LeNet-5](https://github.com/activatedgeek/LeNet-5) on [MNIST](https://pytorch.org/docs/stable/torchvision/datasets.html#mnist)

### Local running environment
For instructions on how to get nuclio and kubernetes up and running locally see [installInstructions.md](https://github.com/Frans-Lukas/DLScheduler/blob/main/InstallInstructions.md) or [Getting Started with Nuclio on Minikube](https://nuclio.io/docs/latest/setup/minikube/getting-started-minikube/).

For instructions on installing and running hadoop/hdfs locally, see [hadoopi.md](https://github.com/Frans-Lukas/DLScheduler/blob/main/hadoopi.md).

### Python
Uses python 3.x

For communcations with hdfs in python see [hdfsPython.md](https://github.com/Frans-Lukas/DLScheduler/blob/main/hdfsPython.md)

### Deploy a nuclio function
To deploy a nuclio functioin see [functionDeploy.md](https://github.com/Frans-Lukas/DLScheduler/blob/main/functionDeploy.md).

