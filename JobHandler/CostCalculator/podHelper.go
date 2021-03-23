package CostCalculator

import (
	v12 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func PodCpuUsage(pod v1beta1.PodMetrics) float64 {
	totalCPUUsed := float64(0)
	for _, container := range pod.Containers {
		for metric, value := range container.Usage {
			if metric == "cpu" {
				val := value.AsApproximateFloat64()
				totalCPUUsed += val
			}
		}
	}
	return totalCPUUsed
}

func PodMemoryUsage(pod v12.Pod) float64 {
	totalMemoryAllocated := float64(0)
	for _, container := range pod.Spec.Containers {
		// TODO decide if we should go by 'limit' or 'request'
		totalMemoryAllocated += container.Resources.Limits.Memory().AsApproximateFloat64()
	}
	return totalMemoryAllocated
}