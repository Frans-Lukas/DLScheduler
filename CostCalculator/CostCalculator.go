package main

import (
	"context"
	"errors"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"os"
	"strings"
	"time"
)

var clientSet *kubernetes.Clientset
var metricsClientSet *metricsv.Clientset

var costPerSec = 0.01
var fixedCost = 0.1
var memoryCost = 0.1

func main() {

	var config string

	if len(os.Args) < 1 {
		log.Fatalf("wrong arguments, REQUIRED <pod-prefix> OPTIONAL FOR OUT OF CLUSTER ACCESS <pathToCfg>")
	}

	prefix := os.Args[1]

	if len(os.Args) > 2 {
		config = os.Args[2] // out of cluster access
	} else {
		config = ""         // in cluster access
	}

	var err error
	err = initializeClients(config)

	fatalErrorCheck(err, "main")

	pods := getNuclioPods()
	metricPods := getNuclioPodsMetrics()

	currentTime := time.Now()
	cost := getTotalCostOfPods(prefix, pods, metricPods, currentTime)

	fatalErrorCheck(err, "main")
	println("Cost is: ", int(cost))
}

func getTotalCostOfPods(prefix string, pods *v12.PodList, metricPods *v1beta1.PodMetricsList, currentTime time.Time) float64 {
	cost := float64(0)

	for _, pod := range pods.Items {
		if matchesPrefix(prefix, pod) {
			memoryPodIndex, err := findMetricsPodIndex(pod.Name, metricPods)
			fatalErrorCheck(err, "getTotalCostOfPods")
			println("pod: ", pod.Name)
			cost += getTotalCostOfpod(pod, metricPods.Items[memoryPodIndex], currentTime)
		}

	}

	return cost
}

func matchesPrefix(prefix string, pod v12.Pod) bool {
	return strings.HasPrefix(pod.Name, prefix)
}

func findMetricsPodIndex(name string, pods *v1beta1.PodMetricsList) (int, error) {
	for i, metricPod := range pods.Items {
		if metricPod.Name == name {
			return i, nil
		}
	}
	return -1, errors.New("findMetricsPod: could not find corresponding metric pod")
}

func getTotalCostOfpod(pod v12.Pod, metricsPod v1beta1.PodMetrics, currentTime time.Time) float64 {
	return fixedCost + getDurationCostOfpod(pod, currentTime) + getMemoryCostOfPod(metricsPod) + getMemoryTransferCostOfpod(pod)
}

func getMemoryCostOfPod(pod v1beta1.PodMetrics) float64 {
	totalMemoryUsed := int64(0)
	for _, container := range pod.Containers {
		for metric, value := range container.Usage {
			if metric == "memory" {
				val, _ := value.AsInt64()
				totalMemoryUsed += val
			}
		}
	}
	println("  memoryCost: ", int64(float64(totalMemoryUsed) * memoryCost))
	return float64(totalMemoryUsed) * memoryCost
}

func getDurationCostOfpod(pod v12.Pod, currentTime time.Time) float64 {
	println("  durationCost: ", int64(currentTime.Sub(pod.CreationTimestamp.Time).Seconds() * costPerSec))
	return currentTime.Sub(pod.CreationTimestamp.Time).Seconds() * costPerSec
}

func getMemoryTransferCostOfpod(pod v12.Pod) float64 {
	return 0 //TODO do this for real
}

func initializeClients(pathToCfg string) error {
	var config *rest.Config
	var err error
	if pathToCfg == "" {
		config, err = rest.InClusterConfig()
		// in cluster access
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", pathToCfg)
	}

	fatalErrorCheck(err, "initializeClients")

	clientSet, err = kubernetes.NewForConfig(config)

	fatalErrorCheck(err, "initializeClients, clientSet")

	metricsClientSet, err = metricsv.NewForConfig(config)

	fatalErrorCheck(err, "initializeClients, metricsClientSet")

	return nil
}

func getNuclioPods() *v12.PodList {
	pods, err := clientSet.CoreV1().Pods("nuclio").List(context.Background(), v1.ListOptions{})

	fatalErrorCheck(err, "getNuclioPods")

	return pods
}

func getNuclioPodsMetrics() *v1beta1.PodMetricsList {
	podMetricsList, err := metricsClientSet.MetricsV1beta1().PodMetricses("nuclio").List(context.Background(), v1.ListOptions{})

	fatalErrorCheck(err, "getNuclioPodMetrics")

	return podMetricsList
}

func fatalErrorCheck(err error, errorPrefix string) {
	if err != nil {
		log.Fatalf(errorPrefix + ": " + err.Error())
	}
}