package main

import (
	"context"
	"jobHandler/helperFunctions"
	"jobHandler/structs"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"strconv"
)

var clientSet *kubernetes.Clientset

func main() {

	// 1. receive job
	if len(os.Args) < 3 {
		log.Fatalf("wrong input, needs arguments <jobPath> and <pathToCfg>")
	}

	err := initializeClients(os.Args[2])
	helperFunctions.FatalErrCheck(err, "main: ")

	// 2. Parse to Job Class
	jobPath := os.Args[1]
	job, err := structs.ParseJson(jobPath)
	helperFunctions.FatalErrCheck(err, "main: ")

	// 3. If done, store gradients and remove job from queue.
	if job.IsDone() {
		println("job is done")
	}

	// 4. Calculate number of functions we want to invoke
	desiredNumberOfFunctions := job.CalculateNumberOfFunctions()

	// 5. Calculate number of functions we can invoke
	numberOfFunctionsToDeploy := deployableNumberOfFunctions(job, desiredNumberOfFunctions)
	println(numberOfFunctionsToDeploy)

	// 6. Invoke functions asynchronously
	deployFunctions(job, numberOfFunctionsToDeploy)

	// 7. Await response from all invoked functions (loss)
	awaitResponse()

	// 8. Save history, and repeat from step 3.
}

func awaitResponse() {

}

func deployFunctions(job structs.Job, numberOfFunctionsToDeploy uint) {
	for i := 0; i < int(numberOfFunctionsToDeploy); i++ {
		deployFunction("job_" + "id_" + strconv.Itoa(i), job.ImageUrl)
	}
}

func deployFunction(podName string, imageUrl string) {
	//TODO
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

	helperFunctions.FatalErrCheck(err, "initializeClients")

	clientSet, err = kubernetes.NewForConfig(config)

	helperFunctions.FatalErrCheck(err, "initializeClients, clientSet")

	return nil
}

func deployableNumberOfFunctions(job structs.Job, desiredNumberOfFunctions uint) uint {
	nodes, err := clientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	helperFunctions.FatalErrCheck(err, "deployableNumberOfFunctions: ")
	if len(nodes.Items) * 2 < int(desiredNumberOfFunctions) {
		return uint(len(nodes.Items) * 2)
	} else {
		return desiredNumberOfFunctions
	}
}
