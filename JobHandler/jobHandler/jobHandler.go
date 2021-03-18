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
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type JobHandler struct {
	ClientSet        *kubernetes.Clientset
	InstancesPerJob  *map[string]uint
	MetricsClientSet *metricsv.Clientset
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

//func (jobHandler JobHandler) InvokeAggregator(job Job, numFunctions uint) {
//	println("running aggregator")
//	functionName := jobHandler.GetPodName(job, 0)
//	jobType := constants.AGGREGATE_JOB_TYPE
//
//	var out bytes.Buffer
//	for {
//		out, stderr, err := helperFunctions.ExecuteFunction(constants.INVOKE_FUNCTION_SCRIPT, functionName, strconv.Itoa(0), strconv.Itoa(int(numFunctions)), jobType)
//
//		helperFunctions.FatalErrCheck(err, "InvokeAggregator: "+out.String()+"\n"+stderr.String())
//		if !strings.Contains(out.String(), "503") && !strings.Contains(out.String(), "500") {
//			break
//		} else {
//			println("503 service unavailable error for aggregator")
//			time.Sleep(2*time.Second)
//		}
//	}
//	println(out.String())
//	println("completed aggregation")
//}

func (jobHandler JobHandler) InvokeFunctions(job Job, numberOfFunctionsToInvoke int) {
	var wg sync.WaitGroup
	numWorkers := job.NumberOfWorkers
	numServers := job.NumberOfServers
	//invoke scheduler
	wg.Add(1)
	functionName := jobHandler.GetPodName(job, 0, constants.JOB_TYPE_SCHEDULER)
	job.PodNames[functionName] = false
	go jobHandler.InvokeWGFunctions(job, 0, *job.Epoch, constants.JOB_TYPE_SCHEDULER, numWorkers, numServers, &wg)
	for i := 0; i < job.NumberOfServers; i++ {
		//invoke servers
		functionName = jobHandler.GetPodName(job, i, constants.JOB_TYPE_SERVER)
		job.PodNames[functionName] = false
		wg.Add(1)
		go jobHandler.InvokeWGFunctions(job, i, *job.Epoch, constants.JOB_TYPE_SERVER, numWorkers, numServers,  &wg)
	}
	//invoke workers
	for i := 0; i < job.NumberOfWorkers; i++ {
		functionName = jobHandler.GetPodName(job, i, constants.JOB_TYPE_WORKER)
		job.PodNames[functionName] = false
		wg.Add(1)
		go jobHandler.InvokeWGFunctions(job, i, *job.Epoch, constants.JOB_TYPE_WORKER, numWorkers, numServers,  &wg)
	}
	//wait for all to complete
	wg.Wait()
}

func (jobHandler JobHandler) InvokeWGFunctions(job Job, id int, epoch int, jobType string, numWorkers int, numServers int, wg *sync.WaitGroup)  {
	defer wg.Done()
	jobHandler.InvokeFunction(job, id, epoch, jobType, numWorkers, numServers)
}

func (jobHandler JobHandler) InvokeFunction(job Job, id int, epoch int, jobType string, numWorkers int, numServers int) {
	println("running function: ", id)
	start := time.Now()
	functionName := jobHandler.GetPodName(job, id, jobType)
	schedulerIp := *job.SchedulerIp
	var response FunctionResponse
	for {
		//'{"ip": "'$2'", "role": "'$3'", "num_workers": '$4', "num_servers": '$5'}'
		out, stderr, err := helperFunctions.ExecuteFunction(constants.INVOKE_FUNCTION_SCRIPT,
			functionName, schedulerIp, jobType, strconv.Itoa(numWorkers), strconv.Itoa(numServers))
		helperFunctions.NonFatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())
		//println(out.String())

		if jobType == constants.JOB_TYPE_WORKER {
			findJson := regexp.MustCompile("regexpresultstart(.*)regexpresultend")
			tmpResponseBody := findJson.FindSubmatch(out.Bytes())
			var responseBody []byte
			if len(tmpResponseBody) > 1 {
				responseBody = tmpResponseBody[1]
			} else {
				println(out.String())
				break
			}
			println(out.String())
			println(responseBody)
			err = json.Unmarshal(responseBody, &response)
			helperFunctions.NonFatalErrCheck(err, "InvokeFunction, regexp: ")
		}
		if strings.Contains(out.String(), "500"){
			err = errors.New("500 internal server error")
		}
		if err == nil {
			break
		} else {
			time.Sleep(time.Second * 3)
		}
	}

	if jobType == constants.JOB_TYPE_WORKER {
		fmt.Println("got response: ", response)
		println("job length: ", len(*job.History))
		*job.History = append(*job.History, HistoryEvent{
			NumWorkers: uint(numWorkers),
			NumServers: uint(numServers),
			WorkerId:   response.WorkerId,
			Loss:       response.Loss,
			Accuracy:   response.Accuracy,
			Time:       time.Since(start).Seconds(),
			Epoch:      epoch,
		})
		println("job length after: ", len(*job.History))
	}
	//println("completed function: ", id)

	//*job.FunctionChannel <- id
}

func (jobHandler JobHandler) AwaitResponse(job Job) {
	for !job.FunctionsHaveFinished() {
		//TODO: fault tolerance, do not allow infinite loop if a function does not return.
		completedFunctionId := jobHandler.GetCompletedFunctionId(job)
		println("function completed with id: ", completedFunctionId)
		job.PodNames[completedFunctionId] = true
	}
}

func (jobHandler JobHandler) GetCompletedFunctionId(job Job) string {
	return <-*job.FunctionChannel
}

func (jobHandler JobHandler) DeployFunctions(job Job) {
	// deploy scheduler
	// deploy x workers
	// deploy y servers
	finishedChannel := make(chan string)

	go jobHandler.DeployChannelFunction(job, 0, finishedChannel, constants.JOB_TYPE_SCHEDULER)

	for i := 0; i < job.NumberOfWorkers; i++ {
		go jobHandler.DeployChannelFunction(job, i, finishedChannel, constants.JOB_TYPE_WORKER)
	}
	for i := 0; i < job.NumberOfServers; i++ {
		go jobHandler.DeployChannelFunction(job, i, finishedChannel, constants.JOB_TYPE_SERVER)
	}
	for i := 0; i < job.NumberOfServers+job.NumberOfWorkers+1; i++ {
		println("pod with id: ", <-finishedChannel, " deployed")
	}
}
func (jobHandler JobHandler) DeployChannelFunction(job Job, functionId int, channel chan string, jobType string) {
	jobHandler.DeployFunction(job, functionId, jobType)
	channel <- jobHandler.GetPodName(job, functionId, jobType)
}

func (jobHandler JobHandler) DeployFunction(job Job, functionId int, jobType string) {
	podName := jobHandler.GetPodName(job, functionId, jobType)
	if jobHandler.podExists(podName){
		return
	}
	imageUrl := job.ImageUrl
	println("Deploying function: ", podName, " with imageUrl: ", imageUrl)

	out, stderr, err := jobHandler.executeDeployFunction(podName, imageUrl)

	if strings.Contains(out.String(), "500"){
		err = errors.New("500 internal server error")
	}

	//TODO: this keeps going until no errors occurs
	for err != nil {
		time.Sleep(time.Second * 5)
		helperFunctions.NonFatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())
		out, stderr, err = jobHandler.executeDeployFunction(podName, imageUrl)
	}


	for jobType == constants.JOB_TYPE_SCHEDULER {
		pods, err := jobHandler.ClientSet.CoreV1().Pods(constants.KUBERNETES_NAMESPACE).List(context.Background(), v1.ListOptions{})
		if err == nil {
			for _, v := range pods.Items {
				if strings.Contains(v.Name, podName) {
					println("found scheduler pod, saving ip")
					*job.SchedulerIp = v.Status.PodIP
					log.Println(out.String())
					return
				}
			}
		}
	}
}

func (jobHandler JobHandler) executeDeployFunction(podName string, imageUrl string) (bytes.Buffer, bytes.Buffer, error) {
	return helperFunctions.ExecuteFunction(
		constants.DEPLOY_FUNCTION_SCRIPT,
		podName,
		imageUrl,
	)
}

func (jobHandler JobHandler) GetPodName(job Job, functionId int, jobType string) string {
	return job.JobId + jobType + strconv.Itoa(functionId)
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

	jobHandler.MetricsClientSet, err = metricsv.NewForConfig(config)

	helperFunctions.FatalErrCheck(err, "initializeClients, MetricsClientSet")


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

func (JobHandler JobHandler) GetDeploymentWithHighestMarginalUtility(jobs []Job, maxFunctions []uint) ([]uint, []uint) {
	if len(jobs) != len(maxFunctions) {
		log.Fatalf("GetDeploymentWithHighestMarginalUtility: len(jobs) != len(maxFunctions)")
	}

	for _, job := range jobs {
		job.UpdateMarginalUtilityFunc()
	}

	workerDeployment := make([]uint, len(jobs))
	serverDeployment := make([]uint, len(jobs))

	deploymentFinished := false

	for !deploymentFinished {
		marginalUtilities := make([]float64, len(jobs))
		deploymentType    := make([]byte, len(jobs))

		for i, job := range jobs {
			// so that no deployment has 1 worker and 0 servers, or 0 servers and 1 worker
			if workerDeployment[i] == 0 {
				utility := job.MarginalUtilityCheck(1, 1, 0, 0, maxFunctions[i])
				marginalUtilities[i] = utility
				deploymentType[i]    = 'f'
			} else {
				workerUtility := job.MarginalUtilityCheck(workerDeployment[i] + 1, serverDeployment[i], workerDeployment[i], serverDeployment[i], maxFunctions[i])
				serverUtility := job.MarginalUtilityCheck(workerDeployment[i], serverDeployment[i] + 1, workerDeployment[i], serverDeployment[i], maxFunctions[i])

				if workerUtility >= serverUtility {
					marginalUtilities[i] = workerUtility
					deploymentType[i]    = 'w'
				} else {
					marginalUtilities[i] = serverUtility
					deploymentType[i]    = 's'
				}
			}
		}

		maxUtility := -1.0
		maxUtilityJobIndex := -1
		for i, utility := range marginalUtilities {
			if utility > maxUtility {
				maxUtility = utility
				maxUtilityJobIndex = i
			}
		}

		if maxUtilityJobIndex == -1 {
			deploymentFinished = true
		} else {
			switch deploymentType[maxUtilityJobIndex] {
			case 'w':
				workerDeployment[maxUtilityJobIndex]++
				break
			case 's':
				serverDeployment[maxUtilityJobIndex]++
				break
			case 'f':
				workerDeployment[maxUtilityJobIndex]++
				serverDeployment[maxUtilityJobIndex]++
				break
			}
		}
	}

	return workerDeployment, serverDeployment
}

func (jobHandler JobHandler) WaitForAllWorkerPods(job Job, namespace string, timeout time.Duration) error {
	hasStarted := false
	for !hasStarted {
		hasStarted = true
		for podName, _ := range job.PodNames {
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

func (jobHandler JobHandler) DeleteNuclioFunctionsInJob(job Job, jobType string, numberOf int) {
	fmt.Printf("deleting all funcitons starting with %s and greater or equal to %d\n", job.JobId + jobType, numberOf)
	stdout, stderr, err := helperFunctions.ExecuteFunction(
		constants.DELETE_FUNCTIONS_SUBSTRING_SCRIPT,
		job.JobId + jobType,
		strconv.Itoa(numberOf),
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
	//numberOfWorkers := job.DataSetSize / batchSize
	//jobHandler.DeployFunction(job, 0)
	//jobHandler.InvokeFunction(job, 0, numberOfWorkers, 1)
	//cost := CostCalculator.CalculateCostForPods(job.JobId, jobHandler.ClientSet, jobHandler.MetricsClientSet)
	//job.UpdateAverageFunctionCost(cost)

	return (*job.History)[len(*job.History) - 1].Time
}

func (jobHandler JobHandler) podExists(name string) bool {
	pods, err := jobHandler.ClientSet.CoreV1().Pods(constants.KUBERNETES_NAMESPACE).List(context.Background(), v1.ListOptions{})
	if err == nil {
		for _, v := range pods.Items {
			if strings.Contains(v.Name, name) {
				return true
			}
		}
	}
	return false
}