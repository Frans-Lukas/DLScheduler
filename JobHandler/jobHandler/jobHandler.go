package jobHandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"jobHandler/constants"
	"jobHandler/helperFunctions"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type JobHandler struct {
	ClientSet *kubernetes.Clientset
	InstancesPerJob *map[string]uint
}

func CreateJobHandler(pthToCfg string) JobHandler {
	var handler JobHandler

	instancesPerJob := make(map[string]uint, 0)
	handler.InstancesPerJob = &instancesPerJob

	err := handler.InitializeClients(pthToCfg)
	println("handler ClientSet: ", handler.ClientSet)

	helperFunctions.FatalErrCheck(err, "CreateJobHandler: ")

	return handler
}

func (jobHandler JobHandler) InvokeAggregator(job Job, numFunctions uint) {
	println("running aggregator")
	functionName := jobHandler.GetPodName(job, 0)
	jobType := constants.AGGREGATE_JOB_TYPE

	var out bytes.Buffer
	for {
		out, stderr, err := helperFunctions.ExecuteFunction(constants.INVOKE_FUNCTION_SCRIPT, functionName, strconv.Itoa(0), strconv.Itoa(int(numFunctions)), jobType)

		helperFunctions.FatalErrCheck(err, "InvokeAggregator: "+out.String()+"\n"+stderr.String())
		if !strings.Contains(out.String(), "503") && !strings.Contains(out.String(), "500") {
			break
		} else {
			println("503 service unavailable error for aggregator")
			time.Sleep(2*time.Second)
		}
	}
	println(out.String())
	println("completed aggregation")
}

func (jobHandler JobHandler) InvokeFunctions(job Job, numberOfFunctionsToInvoke int) {
	var wg sync.WaitGroup
	for i := 0; i < numberOfFunctionsToInvoke; i++ {
		wg.Add(1)
		go jobHandler.InvokeWGFunctions(job, i, numberOfFunctionsToInvoke, *job.Epoch, &wg)
	}
	wg.Wait()
}

func (jobHandler JobHandler) InvokeWGFunctions(job Job, id int, maxId int, epoch int, wg *sync.WaitGroup)  {
	defer wg.Done()
	jobHandler.InvokeFunction(job, id, maxId, epoch)
}

func (jobHandler JobHandler) InvokeFunction(job Job, id int, maxId int, epoch int) {
	println("running function: ", id)
	job.FunctionIds[id] = false
	start := time.Now()
	functionName := jobHandler.GetPodName(job, id)

	var response FunctionResponse
	for {
		out, stderr, err := helperFunctions.ExecuteFunction(constants.INVOKE_FUNCTION_SCRIPT,
			functionName, strconv.Itoa(id), strconv.Itoa(maxId), constants.TRAIN_JOB_TYPE)
		helperFunctions.NonFatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())
		//println(out.String())

		findResponseBody := regexp.MustCompile("Response body:.*\n.*")
		findJson := regexp.MustCompile("\\{(.*)\\}")
		responseBody := findJson.Find(findResponseBody.Find(out.Bytes()))

		println(out.String())
		err = json.Unmarshal(responseBody, &response)
		helperFunctions.NonFatalErrCheck(err, "InvokeFunction, regexp: ")

		if strings.Contains(out.String(), "500"){
			err = errors.New("500 internal server error")
		}
		if err == nil {
			break
		} else {
			time.Sleep(time.Second * 3)
		}
	}
	fmt.Println("got response: ", response)


	println("job length: ", len(*job.History))
	*job.History = append(*job.History, HistoryEvent{
		NumWorkers: uint(maxId),
		WorkerId:   response.WorkerId,
		Loss:       response.Loss,
		Accuracy: 	response.Accuracy,
		Time:       time.Since(start).Seconds(),
		Epoch:      epoch,
	})
	println("job length after: ", len(*job.History))
	//println("completed function: ", id)

	//*job.FunctionChannel <- id
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
		go jobHandler.DeployChannelFunction(job, i, finishedChannel)
	}
	for i := 0; i < int(numberOfFunctionsToDeploy); i++ {
		println("pod with id: ", <-finishedChannel, " deployed")
	}
}
func (jobHandler JobHandler) DeployChannelFunction(job Job, functionId int, channel chan int) {
	jobHandler.DeployFunction(job, functionId)
	channel <- functionId
}

func (jobHandler JobHandler) DeployFunction(job Job, functionId int) {
	//TODO
	imageUrl := job.ImageUrl
	podName := jobHandler.GetPodName(job, functionId)
	println("Deploying function: ", podName, " with imageUrl: ", imageUrl)

	out, stderr, err := helperFunctions.ExecuteFunction(constants.DEPLOY_FUNCTION_SCRIPT, podName, imageUrl)

	if strings.Contains(out.String(), "500"){
		err = errors.New("500 internal server error")
	}

	for err != nil {
		time.Sleep(time.Second * 5)
		helperFunctions.NonFatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())
		out, stderr, err = helperFunctions.ExecuteFunction(constants.DEPLOY_FUNCTION_SCRIPT, podName, imageUrl)
	}

	log.Println(out.String())
}

func (jobHandler JobHandler) GetPodName(job Job, functionId int) string {
	return job.JobId + "job" + strconv.Itoa(functionId)
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

	jobHandler.ClientSet, err = kubernetes.NewForConfig(config)
	println("inside: ", jobHandler.ClientSet)

	helperFunctions.FatalErrCheck(err, "initializeClients, ClientSet")

	return nil
}

func (jobHandler JobHandler) DeployableNumberOfFunctions(job Job, desiredNumberOfFunctions uint) uint {
	nodes, err := jobHandler.ClientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	helperFunctions.FatalErrCheck(err, "deployableNumberOfFunctions: ")
	if len(nodes.Items)*2 < int(desiredNumberOfFunctions) {
		return uint(len(nodes.Items) * 2)
	} else {
		return desiredNumberOfFunctions
	}
}

func (jobHandler JobHandler) WaitForAllWorkerPods(job Job, namespace string, timeout time.Duration) error {
	hasStarted := false
	for !hasStarted {
		hasStarted = true
		for functionId, _ := range job.FunctionIds {
			podName := jobHandler.GetPodName(job, functionId)
			err := helperFunctions.WaitForPodRunning(jobHandler.ClientSet, namespace, podName, timeout)
			if err != nil {
				hasStarted = false
				println(err.Error())
				time.Sleep(time.Second * 2)
				break
			}
		}
	}
	return nil
}

func (jobHandler JobHandler) DeleteNuclioFunctionsInJob(job Job, startRange int, endRange int) {
	stdout, stderr, err := helperFunctions.ExecuteFunction(
		constants.DELETE_FUNCTIONS_SUBSTRING_SCRIPT,
		job.JobId,
		strconv.Itoa(startRange),
		strconv.Itoa(endRange),
	)
	helperFunctions.FatalErrCheck(err, "deleteNuclioFunctionsInJob: "+ stdout.String()+"\n"+stderr.String())
}


func (jobHandler JobHandler) TestReasonableBatchSize(job Job) int {
	datasetSize := job.DataSetSize

	batchSize := 10
	minTimeInSeconds := 50.0
	maxTimeInSeconds := 120.0
	midPointInSeconds := minTimeInSeconds + (maxTimeInSeconds - minTimeInSeconds) // 80
	timeTaken := 0.0

	maxBatchSizeBeforeMinTime := 0
	minBatchSizeBeforeMaxTime := datasetSize

	for timeTaken < minTimeInSeconds || timeTaken > maxTimeInSeconds{
		timeTaken = jobHandler.deployAndRunWithBatchSize(job, batchSize)
		fmt.Printf("took %f seconds with dataset size %d, trying to reach interval %f-%f\n",
			timeTaken, batchSize, minTimeInSeconds, maxTimeInSeconds)

		if timeTaken < minTimeInSeconds {
			if batchSize == datasetSize {
				break
			}
			numTimeTakensToReachInterval := midPointInSeconds / timeTaken
			batchSize = int(math.Min(float64(batchSize)*numTimeTakensToReachInterval, float64(minBatchSizeBeforeMaxTime)))
		} else if timeTaken > maxTimeInSeconds{
			if batchSize == 1 {
				 break
			}
			batchSize = int(math.Max(float64(batchSize) / 2.0, float64(maxBatchSizeBeforeMinTime)))
			if batchSize <= 1 {
				batchSize = 1
			}
		}
	}
	fmt.Printf("Found reasonable batch size %d", batchSize)
	return batchSize
}

func (jobHandler JobHandler) deployAndRunWithBatchSize(job Job, batchSize int) float64 {
	numberOfWorkers := job.DataSetSize / batchSize
	jobHandler.DeployFunction(job, 0)
	jobHandler.InvokeFunction(job, 0, numberOfWorkers, 1)

	return (*job.History)[len(*job.History) - 1].Time
}