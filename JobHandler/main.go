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
	// 1. receive jobs
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
	jobs, err := jb.ParseJson(jobPath)
	helperFunctions.FatalErrCheck(err, "main: ")

	for _, job := range jobs {
		job.JobId = helperFunctions.GenerateId(constants.JOB_ID_LENGTH)
		println("testing reasonable batch size")
		jobHandler.InitialTuning(job)
		println("done with testing reasonable batch size")
	}

	//TODO: check add one one
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.508112, Epoch: 2})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.367166, Epoch: 3})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.327031, Epoch: 4})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.300430, Epoch: 5})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.280054, Epoch: 6})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.262924, Epoch: 7})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.248206, Epoch: 8})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.234580, Epoch: 9})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.221567, Epoch: 10})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.209484, Epoch: 11})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.199290, Epoch: 12})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.190342, Epoch: 13})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.180169, Epoch: 14})
	//*jobs.History = append(*jobs.History, jb.HistoryEvent{Loss: 0.171137, Epoch: 15})
	//for i, _ := range *jobs.History {
	//	//v.Loss *= 100
	//	(*jobs.History)[i].Epoch--
	//	//fmt.Printf("%d, %f\n",v.Epoch, v.Loss)
	//}
	//jobs.LeastSquaresTest()

	//jobHandler.TestReasonableBatchSize(jobs)
	//
	//
	for _, job := range jobs {
		job.JobId = helperFunctions.GenerateId(10)
	}

	// 3. If done, store gradients and remove jobs from queue.
	//for !jobs.IsDone() {
	println("train until convergence")
	trainUntilConvergence(jobHandler, jobs)
}

func trainUntilConvergence(handler jb.JobHandler, jobs []*jb.Job) {
	allJobsDone := false

	functionsDeleted := make([]bool, len(jobs))

	for !allJobsDone {
		jobsReadyForDeployment := make([]*jb.Job, 0)
		outsideServers := uint(0)
		outsideWorkers := uint(0)

		allJobsDone = true
		for _, job := range jobs {
			if !job.CheckIsTraining() && !job.IsDone() {
				job.UpdateIsTraining(true)
				jobsReadyForDeployment = append(jobsReadyForDeployment, job)
			} else if !job.IsDone() {
				outsideServers += job.GetNumberOfServers()
				outsideWorkers += job.GetNumberOfWorkers()
			}
		}

		if len(jobsReadyForDeployment) > 0 {
			trainOneEpoch(handler, jobsReadyForDeployment, outsideWorkers, outsideServers)
		}

		allJobsDone = true
		for i, job := range jobs {
			if !job.IsDone() {
				allJobsDone = false
			} else if !functionsDeleted[i] {
				handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_WORKER, 0)
				functionsDeleted[i] = true
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func trainOneEpoch(handler jb.JobHandler, jobs []*jb.Job, outsideWorkers uint, outsideServers uint) {

	maxFuncs := make([]uint, len(jobs))
	for i, job := range jobs {
		// 4. Calculate number of functions we want to invoke
		desiredNumberOfFunctions := job.CalculateNumberOfFunctions()
		maxFuncs[i] = desiredNumberOfFunctions
		fmt.Printf("desired number of funcs: %d\n", desiredNumberOfFunctions)
	}

	// 5. Calculate number of functions we can invoke
	workerDeployment, serverDeployment := handler.GetDeploymentWithHighestMarginalUtility(jobs, maxFuncs, outsideWorkers, outsideServers)

	for i, job := range jobs {
		//numberOfFunctionsToDeploy := handler.DeployableNumberOfFunctions(job, desiredNumberOfFunctions)
		numberOfFunctionsToDeploy := workerDeployment[i]
		fmt.Printf("actual number of workers: %d\n", numberOfFunctionsToDeploy)

		numberOfServersToDeploy := serverDeployment[i]
		fmt.Printf("actual number of servers: %d\n", numberOfServersToDeploy)

		job.SetNumberOfWorkers(numberOfFunctionsToDeploy)
		job.SetNumberOfServers(numberOfServersToDeploy)

		deleteExcessWorkers(handler, job)
		deleteExcessParameterServers(handler, job)
	}

	//TODO should deployment be threaded?
	for _, job := range jobs {
		// redploy all outsideWorkers and servers, if they exist, they are kept and not redeployed.
		handler.DeployFunctions(job)
	}

	for _, job := range jobs {
		go waitAndExecuteEpochTraining(handler, job)
	}

	// TODO check if this works
	//handler.DeleteNuclioFunctionsInJob(job)
	//if we do not include epoch in pod name we will have to wait for them to delete
}

func waitAndExecuteEpochTraining(handler jb.JobHandler, job *jb.Job) {
	// TODO: wait until function is fully ready before invoking, sleep as a temp solution.
	deployedPods, err := handler.WaitForAllWorkerPods(job, "nuclio", time.Second*10)
	job.DeployedPods = deployedPods
	helperFunctions.FatalErrCheck(err, "waitForAllWorkerPods")

	executeTrainingOfOneEpoch(handler, job)

	job.UpdateIsTraining(false)
}

func deleteExcessWorkers(handler jb.JobHandler, job *jb.Job) {
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_WORKER, job.GetNumberOfWorkers())
}

func deleteExcessParameterServers(handler jb.JobHandler, job *jb.Job) {
	handler.DeleteNuclioFunctionsInJob(job, constants.JOB_TYPE_SERVER, job.GetNumberOfServers())
}

func executeTrainingOfOneEpoch(handler jb.JobHandler, job *jb.Job) {
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
