package jobHandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"jobHandler/CostCalculator"
	"jobHandler/constants"
	"jobHandler/helperFunctions"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	rand2 "k8s.io/apimachinery/pkg/util/rand"
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
	cm               *helperFunctions.ClusterManager
}

func CreateJobHandler(pthToCfg string) JobHandler {
	var handler JobHandler

	instancesPerJob := make(map[string]uint, 0)
	handler.InstancesPerJob = &instancesPerJob

	err := handler.InitializeClients(pthToCfg)
	println("handler ClientSet: ", handler.ClientSet)

	helperFunctions.FatalErrCheck(err, "CreateJobHandler: ")

	//TODO these numbers are nonsense
	//cm := helperFunctions.CreateClusterManager(handler.ClientSet, handler.MetricsClientSet, constants.MAX_SERVERS_PER_NODE, constants.MAX_WORKERS_PER_NODE)
	//handler.cm = &cm
	//handler.cm.UpdateClusterInfo()

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

func (jobHandler JobHandler) InvokeFunctions(job *Job) {
	var wg sync.WaitGroup

	//numWorkers, numServers := jobHandler.countServersAndWorkers(job)
	//
	//job.SetNumberOfServers(numServers)
	//job.SetNumberOfWorkers(numWorkers)

	if job.NumberOfWorkers == 0 && job.NumberOfServers == 0 {
		println("No workers or servers for job: ", job.JobId)
		return
	}

	fmt.Printf("num servers: %d, num workers %d\n", job.NumberOfServers, job.NumberOfWorkers)

	podName := jobHandler.GetPodName(job, 0, constants.JOB_TYPE_SCHEDULER)
	wg.Add(1)
	(*job.PodNames)[podName] = false
	println("invoking ", constants.JOB_TYPE_SCHEDULER)
	go jobHandler.InvokeWGFunction(job, podName, *job.Epoch, constants.JOB_TYPE_SCHEDULER, &wg)

	for i := 0; i < int(job.NumberOfWorkers); i++ {
		podName := jobHandler.GetPodName(job, i, constants.JOB_TYPE_WORKER)
		wg.Add(1)
		(*job.PodNames)[podName] = false
		println("invoking ", constants.JOB_TYPE_WORKER)
		go jobHandler.InvokeWGFunction(job, podName, *job.Epoch, constants.JOB_TYPE_WORKER, &wg)
	}

	for i := 0; i < int(job.NumberOfServers); i++ {
		podName := jobHandler.GetPodName(job, i, constants.JOB_TYPE_SERVER)
		wg.Add(1)
		(*job.PodNames)[podName] = false
		println("invoking ", constants.JOB_TYPE_SERVER)
		go jobHandler.InvokeWGFunction(job, podName, *job.Epoch, constants.JOB_TYPE_SERVER, &wg)
	}

	//for _, podName := range job.DeployedPods {
	//	jobType := parseJobType(podName)
	//	wg.Add(1)
	//	job.PodNames[podName] = false
	//	println("invoking ", jobType)
	//	go jobHandler.InvokeWGFunction(job, podName, *job.Epoch, jobType, &wg)
	//}

	//wait for all to complete
	wg.Wait()
}

func (jobHandler JobHandler) countServersAndWorkers(job *Job) (uint, uint) {
	numWorkers := uint(0)
	numServers := uint(0)
	for _, podName := range job.DeployedPods {
		jobType := parseJobType(podName)
		switch jobType {
		case constants.JOB_TYPE_SERVER:
			numServers++
		case constants.JOB_TYPE_WORKER:
			numWorkers++
		}
	}
	return numWorkers, numServers
}

func (jobHandler JobHandler) InvokeWGFunction(job *Job, id string, epoch int, jobType string, wg *sync.WaitGroup) {
	defer wg.Done()
	jobHandler.InvokeFunction(job, id, epoch, jobType)
	println("function " + id + jobType + " finished.")
}

func (jobHandler JobHandler) InvokeFunction(job *Job, id string, epoch int, jobType string) {
	println("running function: ", id)
	start := time.Now()
	schedulerIp := *job.SchedulerIp
	var response FunctionResponse
	numWorkers := job.NumberOfWorkers
	numServers := job.NumberOfServers
	for {
		//'{"ip": "'$2'", "role": "'$3'", "num_workers": '$4', "num_servers": '$5'}'
		var out, stderr bytes.Buffer
		var err error
		if *job.InitialTuning {
			out, stderr, err = helperFunctions.ExecuteFunction(constants.INVOKE_FUNCTION_SCRIPT,
				id, schedulerIp, jobType, strconv.Itoa(int(numWorkers)),
				strconv.Itoa(int(numServers)), job.ScriptPath, job.JobId, strconv.Itoa(job.NumberOfParts))
		} else {
			out, stderr, err = helperFunctions.ExecuteFunction(constants.INVOKE_FUNCTION_SCRIPT,
				id, schedulerIp, jobType, strconv.Itoa(int(numWorkers)), strconv.Itoa(int(numServers)), job.ScriptPath, job.JobId)
		}
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
		if strings.Contains(out.String(), " 500 "){
			err = errors.New("500 internal server error")
		}
		if err == nil {
			break
		} else {
			println(out.String())
			println(err.Error())
			time.Sleep(time.Second * 3)
		}
	}

	//TODO does this only add history events for workers and not for parameter servers?
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
			Cost:       -1, // is set to real number later
			ActualTrainingEpoch: *job.ActualTrainingStarted,
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
		(*job.PodNames)[completedFunctionId] = true
	}
}

func (jobHandler JobHandler) GetCompletedFunctionId(job Job) string {
	return <-*job.FunctionChannel
}

func (jobHandler JobHandler) DeployFunctions(job *Job) {
	// deploy scheduler
	// deploy x workers
	// deploy y servers
	finishedChannel := make(chan string)

	numServers := job.GetNumberOfServers()
	numWorkers := job.GetNumberOfWorkers()

	if numServers == 0 && numWorkers == 0 {
		return // so that the scheduler isnt deployed
	}

	go jobHandler.DeployChannelFunction(job, 0, finishedChannel, constants.JOB_TYPE_SCHEDULER)

	for i := 0; i < int(numWorkers); i++ {
		go jobHandler.DeployChannelFunction(job, i, finishedChannel, constants.JOB_TYPE_WORKER)
	}
	for i := 0; i < int(numServers); i++ {
		go jobHandler.DeployChannelFunction(job, i, finishedChannel, constants.JOB_TYPE_SERVER)
	}
	for i := 0; i < int(numServers+numWorkers+1); i++ {
		println("pod with id: ", <-finishedChannel, " deployed")
	}
}
func (jobHandler JobHandler) DeployChannelFunction(job *Job, functionId int, channel chan string, jobType string) {
	jobHandler.DeployFunction(job, functionId, jobType)
	channel <- jobHandler.GetPodName(job, functionId, jobType)
}

func (jobHandler JobHandler) DeployFunction(job *Job, functionId int, jobType string) {
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

	for err != nil {
		time.Sleep(time.Second * 5)
		helperFunctions.NonFatalErrCheck(err, "deployFunctions: "+out.String()+"\n"+stderr.String())
		out, stderr, err = jobHandler.executeDeployFunction(podName, imageUrl)
	}

	(*job.PodNames)[podName] = false

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

func (jobHandler JobHandler) GetPodName(job *Job, functionId int, jobType string) string {
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

func (JobHandler JobHandler) GetDeploymentWithHighestMarginalUtility(jobs []*Job, budgets []float64, outsideWorkers uint, outsideServers uint) ([]uint, []uint) {
	if len(jobs) != len(budgets) {
		log.Fatalf("GetDeploymentWithHighestMarginalUtility: len(jobs) != len(budgets)")
	}

	workerDeployment := make([]uint, len(jobs))
	serverDeployment := make([]uint, len(jobs))

	// So that we take into account already existing workers/servers
	workerDeploymentTotal := outsideWorkers
	serverDeploymentTotal := outsideServers

	staticWorkerSetup := make([]bool, len(jobs))
	staticServerSetup := make([]bool, len(jobs))

	println("GetDeploymentWithHighestMarginalUtility: ")

	//JobHandler.cm.UpdateClusterInfo()
	for i, job := range jobs {
		job.UpdateCostFunc()
		job.UpdateMarginalUtilityFunc()
		staticWorkers := job.testingErrors.GetError("staticWorker")
		staticServers := job.testingErrors.GetError("staticServers")

		println(staticWorkers)
		println(staticServers)

		if staticWorkers >= 1 {
			workerDeployment[i]  = uint(staticWorkers)
			staticWorkerSetup[i] = true
		} else {
			staticWorkerSetup[i] = false
		}

		if staticServers >= 1 {
			serverDeployment[i]  = uint(staticServers)
			staticServerSetup[i] = true
		} else {
			staticServerSetup[i] = false
		}
	}

	deploymentFinished := false

	for !deploymentFinished {
		marginalUtilities := make([]float64, len(jobs))
		deploymentType    := make([]byte, len(jobs))

		for i, job := range jobs {
			if workerDeployment[i] == 0 && serverDeployment[i] == 0 {
				utility := -1.0
				// so that no deployment has 1 worker and 0 servers, or 0 servers and 1 worker
				utility = job.MarginalUtilityCheck(1, 1, 0, 0, budgets[i])
				println("\t", job.JobId, " w: ", 1, " s: ", 1, " utility: ", utility)

				marginalUtilities[i] = utility
				deploymentType[i]    = 'f'
			} else {
				workerUtility := -1.0
				workerUtility = job.MarginalUtilityCheck(workerDeployment[i] + 1, serverDeployment[i], workerDeployment[i], serverDeployment[i], budgets[i])
				println("\t", job.JobId, " w: ", workerDeployment[i] + 1, " s: ", serverDeployment[i], " utility: ", workerUtility)

				serverUtility := -1.0
				serverUtility = job.MarginalUtilityCheck(workerDeployment[i], serverDeployment[i] + 1, workerDeployment[i], serverDeployment[i], budgets[i])
				println("\t", job.JobId, " w: ", workerDeployment[i], " s: ", serverDeployment[i] + 1, " utility: ", serverUtility)

				if workerUtility >= serverUtility {
					marginalUtilities[i] = workerUtility
					deploymentType[i]    = 'w'
				} else {
					marginalUtilities[i] = serverUtility
					deploymentType[i]    = 's'
				}
			}
		}

		maxUtility := 0.0
		maxUtilityJobIndex := -1
		for i, utility := range marginalUtilities {
			if utility > maxUtility {
				maxUtility = utility
				maxUtilityJobIndex = i
			}
		}

		if maxUtilityJobIndex == -1 {
			println("GetDeploymentWithHighestMarginalUtility: done!")
			deploymentFinished = true
		} else {
			switch deploymentType[maxUtilityJobIndex] {
			case 'w':
				workerDeployment[maxUtilityJobIndex]++
				workerDeploymentTotal++
				println("\tadding worker to job: ", jobs[maxUtilityJobIndex].JobId)
			case 's':
				serverDeployment[maxUtilityJobIndex]++
				serverDeploymentTotal++
				println("\tadding server to job: ", jobs[maxUtilityJobIndex].JobId)
			case 'f':
				workerDeployment[maxUtilityJobIndex]++
				workerDeploymentTotal++
				serverDeployment[maxUtilityJobIndex]++
				serverDeploymentTotal++
				println("\tadding server and worker to job: ", jobs[maxUtilityJobIndex].JobId)
			}
		}
	}

	println("final deployment:")
	for i, job := range jobs {
		println("\tjob: ", job.JobId, " w: ", workerDeployment[i], " s: ", serverDeployment[i])
	}

	return workerDeployment, serverDeployment
}

//TODO:
/**
* This should wait for at least:
* 1. Worker, 1. Scheduler and 1. Server for each job
* When this requirement has been satisfied, but not all requested pods have been scheduled:
* Wait for x seconds and start invocations!
* Also make sure to check how many of each type are already ready and add them to job configuration.
 */
func (jobHandler JobHandler) WaitForAllWorkerPods(job *Job, namespace string, timeout time.Duration) ([]string, error) {
	hasStarted := false


	startedTypes := make(map[string]int)
	startedPods := make([]string, 0)
	startedTypes[constants.JOB_TYPE_SCHEDULER] = 0
	startedTypes[constants.JOB_TYPE_WORKER] = 0
	startedTypes[constants.JOB_TYPE_WORKER] = 0

	timeStart := time.Now()
	for !hasStarted {
		hasStarted = true
		for podName, _ := range *job.PodNames {
			err := helperFunctions.WaitForPodRunning(jobHandler.ClientSet, namespace, podName, timeout)
			if err != nil {
				if time.Now().Sub(timeStart).Seconds() > time.Second.Seconds() * 100 && allTypesStarted(startedTypes) {
					return startedPods, nil
				}
				hasStarted = false
				println(err.Error())
				time.Sleep(time.Second * 2)
				break
			} else {
				jobType := parseJobType(podName)
				startedPods = append(startedPods, podName)
				startedTypes[jobType]++
			}
		}
	}
	return startedPods, nil
}

func parseJobType(name string) string {
	// 10 chars followed by TYPE followed by int
	re := regexp.MustCompile("[a-z0-9]{10}([a-z]*)[0-9]*")
	return re.FindStringSubmatch(name)[1]
}
func parsePodId(name string) int {
	// 10 chars followed by TYPE followed by int
	re := regexp.MustCompile("[a-z0-9]{10}[a-z]*([0-9]*)")
	id := re.FindStringSubmatch(name)[1]
	idInt, _ := strconv.Atoi(id)
	return idInt
}

func allTypesStarted(types map[string]int) bool {
	return types[constants.JOB_TYPE_SCHEDULER] > 0 && types[constants.JOB_TYPE_WORKER] > 0 && types[constants.JOB_TYPE_SERVER] > 0
}

func (jobHandler JobHandler) DeleteNuclioFunctionsInJob(job *Job, jobType string, numberOf uint) {
	fmt.Printf("deleting all funcitons starting with %s and greater or equal to %d\n", job.JobId + jobType, numberOf)
	stdout, stderr, err := helperFunctions.ExecuteFunction(
		constants.DELETE_FUNCTIONS_SUBSTRING_SCRIPT,
		job.JobId + jobType,
		strconv.Itoa(int(numberOf)),
	)

	podsToDelete := make([]string, 0)
	for i := range *job.PodNames {
		podType := parseJobType(i)
		if podType == jobType {
			podId := parsePodId(i)
			if podId >= int(numberOf){
				podsToDelete = append(podsToDelete, i)
			}
		}
	}

	for _, v := range podsToDelete {
		delete(*job.PodNames, v)
	}

	println(stdout.String())
	helperFunctions.FatalErrCheck(err, "deleteNuclioFunctionsInJob: "+ stdout.String()+"\n"+stderr.String())
}


func (jobHandler JobHandler) InitialTuning(job *Job) int {
	*job.InitialTuning = true
	*job.ActualTrainingStarted = false
	datasetSize := job.DataSetSize

	batchSize := int(math.Min(10, float64(datasetSize)))
	minTimeInSeconds := 10.0
	maxTimeInSeconds := 20.0
	midPointInSeconds := minTimeInSeconds + (maxTimeInSeconds - minTimeInSeconds) // 80
	timeTaken := 0.0

	maxBatchSizeBeforeMinTime := 0
	minBatchSizeBeforeMaxTime := datasetSize

	for timeTaken < minTimeInSeconds || timeTaken > maxTimeInSeconds{
		fmt.Printf("deploying and running with %d dataset size\n", batchSize)
		timeTaken = jobHandler.deployAndRunWithBatchSize(job, batchSize)
		fmt.Printf("took %f seconds with dataset size %d, trying to reach interval %f-%f\n",
			timeTaken, batchSize, minTimeInSeconds, maxTimeInSeconds)

		if timeTaken < minTimeInSeconds {
			if batchSize == datasetSize {
				break
			}
			numTimeTakenToReachInterval := midPointInSeconds / timeTaken
			batchSize = int(math.Min(float64(batchSize)*numTimeTakenToReachInterval, float64(minBatchSizeBeforeMaxTime)))
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

func (jobHandler JobHandler) deployAndRunWithBatchSize(job *Job, batchSize int) float64 {
	job.NumberOfParts = job.DataSetSize / batchSize
	job.SetNumberOfWorkers(1)
	job.SetNumberOfServers(1)

	jobHandler.DeployFunctions(job)

	deployedPods, err := jobHandler.WaitForAllWorkerPods(job, "nuclio", time.Second*10)
	job.DeployedPods = deployedPods
	helperFunctions.FatalErrCheck(err, "waitForAllWorkerPods")
	epochStartTime := time.Now()
	jobHandler.InvokeFunctions(job)
	cost := CostCalculator.CalculateCostForPods(job.JobId, jobHandler.ClientSet, jobHandler.MetricsClientSet, epochStartTime)
	job.UpdateFunctionCostsInHistory(cost)
	return (*job.History)[len(*job.History) - 1].Time
}

func (jobHandler JobHandler) RunMiniEpoch(job *Job, batchSize int) {
	*job.ActualTrainingStarted = false

	job.NumberOfParts = job.DataSetSize / batchSize
	job.SetNumberOfWorkers(uint(rand2.IntnRange(1, 4)))
	job.SetNumberOfServers(uint(rand2.IntnRange(1, 4)))


	fmt.Printf("running mini epoch with %d workers and %d servers\n", job.NumberOfWorkers, job.NumberOfServers)

	jobHandler.DeployFunctions(job)

	jobHandler.DeleteExcessWorkers(job)
	jobHandler.DeleteExcessParameterServers(job)

	deployedPods, err := jobHandler.WaitForAllWorkerPods(job, "nuclio", time.Second*10)
	job.DeployedPods = deployedPods
	helperFunctions.FatalErrCheck(err, "waitForAllWorkerPods")
	epochStartTime := time.Now()
	jobHandler.InvokeFunctions(job)
	cost := CostCalculator.CalculateCostForPods(job.JobId, jobHandler.ClientSet, jobHandler.MetricsClientSet, epochStartTime)
	job.UpdateFunctionCostsInHistory(cost)
}


func (handler JobHandler) DeleteExcessWorkers(job *Job) {
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_WORKER, job.GetNumberOfWorkers())
}

func (handler JobHandler) DeleteExcessParameterServers(job *Job) {
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_SERVER, job.GetNumberOfServers())
}

func (jobHandler JobHandler) podExists(name string) bool {
	pods, err := jobHandler.ClientSet.CoreV1().Pods(constants.KUBERNETES_NAMESPACE).List(context.Background(), v1.ListOptions{})
	if err == nil {
		for _, v := range pods.Items {
			if strings.Contains(v.Name, name) {
				println("pod ", name, " exists")
				return true
			}
		}
	}
	return false
}