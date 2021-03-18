package CostCalculator

import (
	"context"
	"errors"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"strings"
	"time"
)

var costPerSec = 0.01
var fixedCost = 0.1
var memoryCost = 0.0001
var CPUCost = 10000.0

func CalculateCostForPods(subString string, clientSet *kubernetes.Clientset, metricsClientSet *metricsv.Clientset, jobStartTime time.Time) float64 {
	pods := getNuclioPods(clientSet)
	metricPods := getNuclioPodsMetrics(metricsClientSet)

	currentTime := time.Now()
	cost := getTotalCostOfPods(subString, pods, metricPods, currentTime, jobStartTime)

	println("Cost is: ", int(cost))

	return cost
}

func getTotalCostOfPods(subString string, pods *v12.PodList, metricPods *v1beta1.PodMetricsList, currentTime time.Time, jobStartTime time.Time) float64 {
	cost := float64(0)

	for _, pod := range pods.Items {
		if matchesSubstring(subString, pod) {
			memoryPodIndex, err := findMetricsPodIndex(pod.Name, metricPods)
			nonFatalErrorCheck(err, "getTotalCostOfPods")
			println("pod: ", pod.Name)
			if memoryPodIndex != -1 {
				cost += getTotalCostOfpod(pod, metricPods.Items[memoryPodIndex], currentTime, jobStartTime)
			} else {
				cost += getTotalCostOfpodWithoutMetricPod(pod, currentTime, jobStartTime)
			}
		}
	}

	return cost
}

func matchesSubstring(subString string, pod v12.Pod) bool {
	return strings.Contains(pod.Name, subString)
}

func findMetricsPodIndex(name string, pods *v1beta1.PodMetricsList) (int, error) {
	for i, metricPod := range pods.Items {
		if metricPod.Name == name {
			return i, nil
		}
	}
	return -1, errors.New("findMetricsPod: could not find corresponding metric pod")
}

func getTotalCostOfpodWithoutMetricPod(pod v12.Pod, currentTime time.Time, jobStartTime time.Time) float64 {
	return fixedCost + getDurationCostOfpod(pod, currentTime, jobStartTime) + getMemoryCostOfPod(pod) + getMemoryTransferCostOfpod(pod)
}

func getTotalCostOfpod(pod v12.Pod, metricsPod v1beta1.PodMetrics, currentTime time.Time, jobStartTime time.Time) float64 {
	return getTotalCostOfpodWithoutMetricPod(pod, currentTime, jobStartTime) + getCPUCostOfPod(metricsPod) + getGPUCostOfPod()
}

func getGPUCostOfPod() float64 {
	//TODO figure out how to do this
	return 0.0
}

func getCPUCostOfPod(pod v1beta1.PodMetrics) float64 {
	totalCPUUsed := float64(0)
	for _, container := range pod.Containers {
		for metric, value := range container.Usage {
			if metric == "cpu" {
				val := value.AsApproximateFloat64()
				totalCPUUsed += val
			}
		}
	}
	println("  CPUCost: ", totalCPUUsed *CPUCost)
	return (totalCPUUsed) * CPUCost
}

func getMemoryCostOfPod(pod v12.Pod) float64 {
	totalMemoryAllocated := float64(0)
	for _, container := range pod.Spec.Containers {
		// TODO decide if we should go by 'limit' or 'request'
		totalMemoryAllocated += container.Resources.Limits.Memory().AsApproximateFloat64()
	}
	println("  memoryCost: ", totalMemoryAllocated *memoryCost)
	return totalMemoryAllocated * memoryCost
}

func getDurationCostOfpod(pod v12.Pod, currentTime time.Time, jobStartTime time.Time) float64 {
	println("  durationCost: ", int64(currentTime.Sub(pod.CreationTimestamp.Time).Seconds() *costPerSec))

	if pod.CreationTimestamp.Time.Before(jobStartTime) {
		return currentTime.Sub(jobStartTime).Seconds() * costPerSec
	} else {
		return currentTime.Sub(pod.CreationTimestamp.Time).Seconds() * costPerSec
	}
}

func getMemoryTransferCostOfpod(pod v12.Pod) float64 {
	return 0 //TODO do this for real
}

func getNuclioPods(clientSet *kubernetes.Clientset) *v12.PodList {
	pods, err := clientSet.CoreV1().Pods("nuclio").List(context.Background(), v1.ListOptions{})

	fatalErrorCheck(err, "getNuclioPods")

	return pods
}

func getNuclioPodsMetrics(metricsClientSet *metricsv.Clientset) *v1beta1.PodMetricsList {
	podMetricsList, err := metricsClientSet.MetricsV1beta1().PodMetricses("nuclio").List(context.Background(), v1.ListOptions{})

	fatalErrorCheck(err, "getNuclioPodMetrics")

	return podMetricsList
}

func nonFatalErrorCheck(err error, errorPrefix string) {
	if err != nil {
		println(errorPrefix + ": " + err.Error())
	}
}

func fatalErrorCheck(err error, errorPrefix string) {
	if err != nil {
		log.Fatalf(errorPrefix + ": " + err.Error())
	}
}