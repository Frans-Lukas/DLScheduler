package helperFunctions

import (
	"context"
	"errors"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type ClusterManager struct {
	clientSet          *kubernetes.Clientset
	activeNodes        *coreV1.NodeList
	metricsSet         *metricsv.Clientset
	activeNodesMetrics *v1beta1.NodeMetricsList
	maxServersPerNode  uint
	maxWorkersPerNode  uint
}

func CreateClusterManager(clientSet *kubernetes.Clientset, metricsSet *metricsv.Clientset, maxServersPerNode uint, maxWorkersPerNode uint) ClusterManager {
	var cm ClusterManager

	cm.clientSet = clientSet
	cm.metricsSet = metricsSet
	cm.maxServersPerNode = maxServersPerNode
	cm.maxWorkersPerNode = maxWorkersPerNode

	return cm
}

func (cm *ClusterManager) UpdateClusterInfo() {
	var err error
	cm.activeNodes, err = cm.clientSet.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
	FatalErrCheck(err, "UpdateClusterInfo: ")

	cm.activeNodesMetrics, err = cm.metricsSet.MetricsV1beta1().NodeMetricses().List(context.Background(), metaV1.ListOptions{})
	FatalErrCheck(err, "UpdateClusterInfo: ")
}

func (cm *ClusterManager) CheckDeploymentValidity(numWorkers uint, numServers uint) bool {
	numberOfNodes := uint(len(cm.activeNodes.Items))

	numberCheck :=  numWorkers <= cm.maxWorkersPerNode * numberOfNodes && numServers <= cm.maxServersPerNode * numberOfNodes

	if !numberCheck {
		return false
	}

	memoryCheck := cm.remainingMemoryCheck(numWorkers, numServers)
	//TODO if we want to use this we should do the same for cpu usage and gpu usage

	if !memoryCheck {
		return false
	}

	return true
}

func (cm *ClusterManager) remainingMemoryCheck(workers uint, servers uint) bool {
	//remainingMemory := cm.remainingMemory()

	//TODO figure out how to do this, if we want to use it
	//neededMemory := memoryNeededForPods(worker, servers)

	//return neededMemory < remainingMemory

	return true
}

func (cm *ClusterManager) remainingMemory() int64 {
	totalMemoryUsage := int64(0)

	for _, node := range cm.activeNodesMetrics.Items {
		memoryUsage, check := node.Usage.Memory().AsInt64()

		if !check {
			FatalErrCheck(errors.New("failed memory reading"), "remainingMemory: ") //TODO figure out what to do here
		}

		totalMemoryUsage += memoryUsage
	}

	totalMemory := int64(0)

	for _, node := range cm.activeNodes.Items {
		memory, check := node.Status.Allocatable.Memory().AsInt64()

		if !check {
			FatalErrCheck(errors.New("failed memory reading"), "remainingMemory: ") //TODO figure out what to do here
		}

		totalMemory += memory
	}

	return totalMemory - totalMemoryUsage
}