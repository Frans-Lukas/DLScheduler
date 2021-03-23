package main

import (
	"fmt"
	"jobHandler/CostCalculator"
	"jobHandler/constants"
	"jobHandler/helperFunctions"
	jb "jobHandler/jobHandler"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	// 1. receive job
	if len(os.Args) < 2 {
		log.Fatalf("wrong input, needs arguments <jobPath> and optional <pathToCfg>")
	}

	var jobHandler jb.JobHandler

	var err error
	if len(os.Args) > 2 {
		jobHandler = jb.CreateJobHandler(os.Args[2])
	} else {
		jobHandler = jb.CreateJobHandler("")
	}

	helperFunctions.FatalErrCheck(err, "main: ")

	// 2. Parse to Job Class
	jobPath := os.Args[1]
	job, err := jb.ParseJson(jobPath)
	helperFunctions.FatalErrCheck(err, "main: ")

	for _,v := range os.Environ() {
		println(v)
	}
	//println(os.Environ())

	//TODO: check add one one
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.508112, Epoch: 2})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.367166, Epoch: 3})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.327031, Epoch: 4})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.300430, Epoch: 5})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.280054, Epoch: 6})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.262924, Epoch: 7})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.248206, Epoch: 8})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.234580, Epoch: 9})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.221567, Epoch: 10})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.209484, Epoch: 11})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.199290, Epoch: 12})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.190342, Epoch: 13})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.180169, Epoch: 14})
	//*job.History = append(*job.History, jb.HistoryEvent{Loss: 0.171137, Epoch: 15})
	//for i, _ := range *job.History {
	//	//v.Loss *= 100
	//	(*job.History)[i].Epoch--
	//	//fmt.Printf("%d, %f\n",v.Epoch, v.Loss)
	//}
	//job.LeastSquaresTest()

	//jobHandler.TestReasonableBatchSize(job)
	//
	//
	job.JobId = helperFunctions.GenerateId(10)

	// 3. If done, store gradients and remove job from queue.
	//for !job.IsDone() {
	println("train until convergence")
	trainUntilConvergence(jobHandler, job)
}

func trainUntilConvergence(handler jb.JobHandler, job jb.Job) {
	for !job.IsDone() {
		// 4. Calculate number of functions we want to invoke
		desiredNumberOfFunctions := job.CalculateNumberOfFunctions()
		fmt.Printf("desired number of funcs: %d\n", desiredNumberOfFunctions)
		// 5. Calculate number of functions we can invoke
		jobs := []jb.Job{job}
		maxFuncs := []uint{desiredNumberOfFunctions}
		workerDeployment, serverDeployment := handler.GetDeploymentWithHighestMarginalUtility(jobs, maxFuncs)

		//numberOfFunctionsToDeploy := handler.DeployableNumberOfFunctions(job, desiredNumberOfFunctions)
		numberOfFunctionsToDeploy := workerDeployment[0]
		fmt.Printf("actual number of workers: %d\n", numberOfFunctionsToDeploy)

		numberOfServersToDeploy := serverDeployment[0]
		fmt.Printf("actual number of servers: %d\n", numberOfServersToDeploy)

		job.NumberOfWorkers = numberOfFunctionsToDeploy
		job.NumberOfServers = numberOfServersToDeploy

		deleteExcessWorkers(handler, job)
		deleteExcessParameterServers(handler, job)

		// redploy all workers and servers, if they exist, they are kept and not redeployed.
		handler.DeployFunctions(job)

		// TODO: wait until function is fully ready before invoking, sleep as a temp solution.
		err := handler.WaitForAllWorkerPods(job, "nuclio", time.Second*10)
		helperFunctions.FatalErrCheck(err, "waitForAllWorkerPods")

		trainOneEpoch(handler, job, numberOfFunctionsToDeploy)

		// TODO check if this works
		//handler.DeleteNuclioFunctionsInJob(job)
		//if we do not include epoch in pod name we will have to wait for them to delete
	}
}

func deleteExcessWorkers(handler jb.JobHandler, job jb.Job) {
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_WORKER, job.NumberOfWorkers)
}

func deleteExcessParameterServers(handler jb.JobHandler, job jb.Job) {
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_SERVER, job.NumberOfServers)
}

func trainOneEpoch(handler jb.JobHandler, job jb.Job, numberOfFunctionsToInvoke uint) {
	println("invoking functions")

	epochStartTime := time.Now()

	handler.InvokeFunctions(job)

	// print history events and loss estimation function
	//job.LeastSquaresTest()

	*job.Epoch++

	// update costs for functions
	cost := CostCalculator.CalculateCostForPods(job.JobId, handler.ClientSet, handler.MetricsClientSet, epochStartTime)
	job.UpdateAverageFunctionCost(cost)

	println("job is done")
}
