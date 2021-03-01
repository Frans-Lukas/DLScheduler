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
	"math/rand"
	"os"
	"strconv"
	"time"
)

var clientSet *kubernetes.Clientset

func main() {
	rand.Seed(time.Now().UnixNano())

	// 1. receive job
	if len(os.Args) < 2 {
		log.Fatalf("wrong input, needs arguments <jobPath> and optional <pathToCfg>")
	}

	var err error
	if len(os.Args) > 2 {
		err = initializeClients(os.Args[2])
	} else {
		err = initializeClients("")
	}

	helperFunctions.FatalErrCheck(err, "main: ")

	// 2. Parse to Job Class
	jobPath := os.Args[1]
	job, err := structs.ParseJson(jobPath)
	helperFunctions.FatalErrCheck(err, "main: ")

	job.JobId = helperFunctions.GenerateId(10)

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
	invokeFunctions(job, numberOfFunctionsToDeploy)

	// 7. Await response from all invoked functions (loss)
	awaitResponse(job)

	// 8. Save history, and repeat from step 3.
}

func invokeFunctions(job structs.Job, numberOfFunctionsToInvoke uint) {
	for i := 0; i < int(numberOfFunctionsToInvoke); i++ {
		invokeFunction(job, i)
	}
}

func invokeFunction(job structs.Job, id int) {
	job.FunctionChannel <- id
	job.History = append(job.History, structs.HistoryEvent{
		NumWorkers: 0,
		Loss:       0,
		Time:       0,
		Epoch:      0,
	})
}

func awaitResponse(job structs.Job) {
	for job.FunctionsHaveFinished() {
		//TODO: fault tolerance, do not allow infinite loop if a function does not return.
		job.FunctionIds[getCompletedFunctionId(job)] = true
	}
}

func getCompletedFunctionId(job structs.Job) int {
	return <-job.FunctionChannel
}

func deployFunctions(job structs.Job, numberOfFunctionsToDeploy uint) {
	for i := 0; i < int(numberOfFunctionsToDeploy); i++ {
		deployFunction(job, i)
	}
}

func deployFunction(job structs.Job, functionId int) {
	//TODO
	imageUrl := job.ImageUrl
	podName := "job_"+job.JobId+"_"+strconv.Itoa(functionId)
	println("Deploying function: ", podName, " with imageUrl: ", imageUrl)
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
	if len(nodes.Items)*2 < int(desiredNumberOfFunctions) {
		return uint(len(nodes.Items) * 2)
	} else {
		return desiredNumberOfFunctions
	}
}
