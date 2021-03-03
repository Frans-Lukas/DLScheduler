package helperFunctions

import (
	"context"
	"errors"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

func IsPodRunning(client kubernetes.Interface, podName string, namespace string) wait.ConditionFunc {
	return func() (done bool, err error) {
		pod, err := client.CoreV1().Pods(namespace).Get(context.Background(), podName, metaV1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
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