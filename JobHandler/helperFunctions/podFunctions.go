package helperFunctions

import (
	"context"
	"errors"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

func IsPodRunning(client kubernetes.Interface, podName string, namespace string) wait.ConditionFunc {
	return func() (done bool, err error) {
		podList, err := client.CoreV1().Pods(namespace).List(context.Background(), metaV1.ListOptions{ })
		if err != nil {
			return false, err
		}
		var pod *v1.Pod = nil
		for _, currPod := range podList.Items {
			if strings.Contains(currPod.Name, podName){
				pod = &currPod
				break
			}
		}
		if pod == nil {
			return false, errors.New("podName does not exist: " + podName)
		}


		switch pod.Status.Phase {
		case v1.PodRunning:
			terminating := false
			for _, status := range pod.Status.ContainerStatuses {
				if status.State.Terminated != nil {
					terminating = true
					break
				}
			}
			if terminating {
				return false, errors.New("PodTerminated")
			} else {
				return true, nil
			}
		case v1.PodSucceeded:
			return false, nil
		case v1.PodFailed:
			return false, errors.New("PodFailed")
		}

		return false, nil
	}
}


func WaitForPodRunning(client kubernetes.Interface, namespace, podName string, timeout time.Duration) error {
	return wait.PollImmediate(time.Second, timeout, IsPodRunning(client, podName, namespace))
}
