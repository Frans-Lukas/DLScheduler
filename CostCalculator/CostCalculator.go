package main

import (
	"context"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"time"
)

var clientSet *kubernetes.Clientset

var costPerSec = 0.01
var fixedCost = 0.1

func main() {

	var config string
	if len(os.Args) > 1 {
		config = os.Args[1] // out of cluster access
	} else {
		config = ""         // in cluster access
	}

	var err error
	clientSet, err = getClient(config)

	if err != nil {
		log.Fatalf("main: " + err.Error())
	}

	pods := getNuclioPods()

	currentTime := time.Now()
	cost := getTotalCostOfPods(pods, currentTime)

	println("Cost is: ", cost)
}

func getTotalCostOfPods(pods *v12.PodList, currentTime time.Time) float64 {
	cost := float64(0)

	for _, pod := range pods.Items {
		cost += getTotalCostOfpod(pod, currentTime)
	}

	return cost
}

func getTotalCostOfpod(pod v12.Pod, currentTime time.Time) float64 {
	return fixedCost + getDurationCostOfpod(pod, currentTime) + getMemoryTransferCostOfpod(pod)
}

func getDurationCostOfpod(pod v12.Pod, currentTime time.Time) float64 {
	println(currentTime.Sub(pod.CreationTimestamp.Time).Seconds())
	return currentTime.Sub(pod.CreationTimestamp.Time).Seconds() * costPerSec
}

func getMemoryTransferCostOfpod(pod v12.Pod) float64 {
	return 0 //TODO do this for real
}

func getClient(pathToCfg string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if pathToCfg == "" {
		config, err = rest.InClusterConfig()
		// in cluster access
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", pathToCfg)
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getNuclioPods() *v12.PodList {
	pods, err := clientSet.CoreV1().Pods("nuclio").List(context.Background(), v1.ListOptions{})

	if err != nil {
		log.Fatalf("getpodsByName: " + err.Error())
	}

	return pods
}