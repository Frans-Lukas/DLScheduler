package jobHandler

import (
	"context"
	"encoding/json"
	"fmt"
	"jobHandler/constants"
	"jobHandler/helperFunctions"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"regexp"
	"strconv"
	"time"
)

type JobHandler struct {
	clientSet *kubernetes.Clientset
}

func CreateJobHandler(pthToCfg string) JobHandler {
	var handler JobHandler
	err := handler.InitializeClients(pthToCfg)
	println("handler clientSet: ", handler.clientSet)

	helperFunctions.FatalErrCheck(err, "CreateJobHandler: ")

	return handler
}

func (jobHandler JobHandler) InvokeAggregator(job Job, numFunctions uint) {
	println("running aggregator")
	functionName := jobHandler.GetPodName(job, 0)
	jobType := constants.AGGREGATE_JOB_TYPE

	out, stderr, err := helperFunctions.ExecuteFunction(constants.INVOKE_FUNCTION_SCRIPT, functionName, strconv.Itoa(0), strconv.Itoa(int(numFunctions)), jobType)

	helperFunctions.FatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())
	println(out.String())
	println("completed aggregation")
}

func (jobHandler JobHandler) InvokeFunctions(job Job, numberOfFunctionsToInvoke int) {
	for i := 0; i < numberOfFunctionsToInvoke; i++ {
		go jobHandler.InvokeFunction(job, i, numberOfFunctionsToInvoke)
	}
}

func (jobHandler JobHandler) InvokeFunction(job Job, id int, maxId int) {
	println("running function: ", id)
	start := time.Now()
	functionName := jobHandler.GetPodName(job, id)

	out, stderr, err := helperFunctions.ExecuteFunction(constants.INVOKE_FUNCTION_SCRIPT,
		functionName, strconv.Itoa(id), strconv.Itoa(maxId), constants.TRAIN_JOB_TYPE)

	helperFunctions.FatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())
	//println(out.String())

	findResponseBody := regexp.MustCompile("Response body:.*\n.*")
	findJson := regexp.MustCompile("\\{(.*)\\}")
	responseBody := findJson.Find(findResponseBody.Find(out.Bytes()))

	var response FunctionResponse
	err = json.Unmarshal(responseBody, &response)
	helperFunctions.FatalErrCheck(err, "deployFunctions, regexp: ")
	fmt.Println("got response: ", response)

	*job.FunctionChannel <- id
	job.History = append(job.History, HistoryEvent{
		NumWorkers: uint(maxId),
		WorkerId:   response.WorkerId,
		Loss:       response.Loss,
		Time:       time.Since(start).Seconds(),
		Epoch:      job.Epoch,
	})
	//println("completed function: ", id)
}

func (jobHandler JobHandler) AwaitResponse(job Job) {
	for !job.FunctionsHaveFinished() {
		//TODO: fault tolerance, do not allow infinite loop if a function does not return.
		completedFunctionId := jobHandler.GetCompletedFunctionId(job)
		println("function completed with id: ", completedFunctionId)
		job.FunctionIds[completedFunctionId] = true
	}
}

func (jobHandler JobHandler) GetCompletedFunctionId(job Job) int {
	return <-*job.FunctionChannel
}

func (jobHandler JobHandler) DeployFunctions(job Job, numberOfFunctionsToDeploy uint) {
	var finishedChannel chan int
	finishedChannel = make(chan int)
	for i := 0; i < int(numberOfFunctionsToDeploy); i++ {
		go jobHandler.DeployFunction(job, i, finishedChannel)
	}
	for i := 0; i < int(numberOfFunctionsToDeploy); i++ {
		println("pod with id: ", <-finishedChannel, " deployed")
	}
}

func (jobHandler JobHandler) DeployFunction(job Job, functionId int, channel chan int) {
	//TODO
	imageUrl := job.ImageUrl
	podName := jobHandler.GetPodName(job, functionId)
	println("Deploying function: ", podName, " with imageUrl: ", imageUrl)

	out, stderr, err := helperFunctions.ExecuteFunction(constants.DEPLOY_FUNCTION_SCRIPT, podName, imageUrl)

	helperFunctions.FatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())

	job.FunctionIds[functionId] = false

	log.Println(out.String())

	channel <- functionId
}

func (jobHandler JobHandler) GetPodName(job Job, functionId int) string {
	return "job_" + job.JobId + "_" + strconv.Itoa(functionId)
}

func (jobHandler *JobHandler) InitializeClients(pathToCfg string) error {
	var config *rest.Config
	var err error
	if pathToCfg == "" {
		config, err = rest.InClusterConfig()
		// in cluster access
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", pathToCfg)
	}

	helperFunctions.FatalErrCheck(err, "initializeClients")

	jobHandler.clientSet, err = kubernetes.NewForConfig(config)
	println("inside: ", jobHandler.clientSet)

	helperFunctions.FatalErrCheck(err, "initializeClients, clientSet")

	return nil
}

func (jobHandler JobHandler) DeployableNumberOfFunctions(job Job, desiredNumberOfFunctions uint) uint {
	nodes, err := jobHandler.clientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	helperFunctions.FatalErrCheck(err, "deployableNumberOfFunctions: ")
	if len(nodes.Items)*2 < int(desiredNumberOfFunctions) {
		return uint(len(nodes.Items) * 2)
	} else {
		return desiredNumberOfFunctions
	}
}
