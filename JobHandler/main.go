package main

import (
	"bytes"
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
	"os/exec"
	"strconv"
	"time"
)

var clientSet *kubernetes.Clientset
const DEPLOY_FUNCTION_SCRIPT = "./nuclio/deploy_nuclio_docker_container.sh"
const INVOKE_FUNCTION_SCRIPT = "./nuclio/invoke_nuclio_function.sh"
const TRAIN_JOB_TYPE = "train"
const AGGREGATE_JOB_TYPE = "average"

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
	//for !job.IsDone() {

	// 4. Calculate number of functions we want to invoke
	desiredNumberOfFunctions := job.CalculateNumberOfFunctions()

	// 5. Calculate number of functions we can invoke
	numberOfFunctionsToDeploy := deployableNumberOfFunctions(job, desiredNumberOfFunctions)
	numberOfFunctionsToDeploy = 5
	println(numberOfFunctionsToDeploy)

	// 6. Invoke functions asynchronously
	deployFunctions(job, numberOfFunctionsToDeploy)
	println("invoking functions")
	invokeFunctions(job, int(numberOfFunctionsToDeploy))

	// 7. Await response from all invoked functions (loss)
	println("waiting for invocation responses")
	awaitResponse(job)

	// 8. aggregate history, and repeat from step 3.
	invokeAggregator(job, numberOfFunctionsToDeploy)
	println("job is done")
	//
	//}
}

func invokeAggregator(job structs.Job, numFunctions uint) {
	println("running aggregator")
	functionName := getPodName(job, 0)
	jobType := AGGREGATE_JOB_TYPE

	cmd := exec.Command(INVOKE_FUNCTION_SCRIPT, functionName, strconv.Itoa(0), strconv.Itoa(int(numFunctions)), jobType)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	helperFunctions.FatalErrCheck(err, "deployFunctions: " + out.String() + "\n" + stderr.String())
	println(out.String())
	println("completed aggregation")
}

func invokeFunctions(job structs.Job, numberOfFunctionsToInvoke int) {
	for i := 0; i < numberOfFunctionsToInvoke; i++ {
		go invokeFunction(job, i, numberOfFunctionsToInvoke)
	}
}

func invokeFunction(job structs.Job, id int, maxId int) {
	println("running function: ", id)
	functionName := getPodName(job, id)
	jobType := TRAIN_JOB_TYPE

	cmd := exec.Command(INVOKE_FUNCTION_SCRIPT, functionName, strconv.Itoa(id), strconv.Itoa(maxId), jobType)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	helperFunctions.FatalErrCheck(err, "deployFunctions: " + out.String() + "\n" + stderr.String())

	*job.FunctionChannel <- id
	job.History = append(job.History, structs.HistoryEvent{
		NumWorkers: uint(maxId),
		Loss:       0,
		Time:       0,
		Epoch:      0,
	})
	println(out.String())
	println("completed function: ", id)
}

func awaitResponse(job structs.Job) {
	for !job.FunctionsHaveFinished() {
		//TODO: fault tolerance, do not allow infinite loop if a function does not return.
		completedFunctionId := getCompletedFunctionId(job)
		println("function completed with id: ", completedFunctionId)
		job.FunctionIds[completedFunctionId] = true
	}
}

func getCompletedFunctionId(job structs.Job) int {
	return <-*job.FunctionChannel
}

func deployFunctions(job structs.Job, numberOfFunctionsToDeploy uint) {
	var finishedChannel chan int
	finishedChannel = make(chan int)
	for i := 0; i < int(numberOfFunctionsToDeploy); i++ {
		go deployFunction(job, i, finishedChannel)
	}
	for i := 0; i < int(numberOfFunctionsToDeploy); i++ {
		println("pod with id: ", <- finishedChannel, " deployed")
	}
}

func deployFunction(job structs.Job, functionId int, channel chan int) {
	//TODO
	imageUrl := job.ImageUrl
	podName := getPodName(job, functionId)
	println("Deploying function: ", podName, " with imageUrl: ", imageUrl)

	cmd := exec.Command(DEPLOY_FUNCTION_SCRIPT, podName, imageUrl)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	helperFunctions.FatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())

	job.FunctionIds[functionId] = false

	log.Println(out.String())

	channel <- functionId
}

func getPodName(job structs.Job, functionId int) string {
	return "job_" + job.JobId + "_" + strconv.Itoa(functionId)
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
