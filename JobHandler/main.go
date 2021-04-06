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
		log.Fatalf("wrong input, needs arguments <jobPath> and optional <pathToCfg>, e.x. exampleJob.json /home/franslukas/.kube/config")
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

	job.JobId = helperFunctions.GenerateId(constants.JOB_ID_LENGTH)
	println("testing reasonable batch size")
	jobHandler.InitialTuning(job)
	println("done with testing reasonable batch size")

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

		deployedPods, err := handler.WaitForAllWorkerPods(job, "nuclio", time.Second*10)
		job.DeployedPod = deployedPods
		helperFunctions.FatalErrCheck(err, "waitForAllWorkerPods")

		trainOneEpoch(handler, job)
	}
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_WORKER, 0)
}

func deleteExcessWorkers(handler jb.JobHandler, job jb.Job) {
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_WORKER, job.NumberOfWorkers)
}

func deleteExcessParameterServers(handler jb.JobHandler, job jb.Job) {
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_SERVER, job.NumberOfServers)
}

func trainOneEpoch(handler jb.JobHandler, job jb.Job) {
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
