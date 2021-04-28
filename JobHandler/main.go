package main

import (
	"bufio"
	"fmt"
	"jobHandler/CostCalculator"
	"jobHandler/constants"
	"jobHandler/helperFunctions"
	jb "jobHandler/jobHandler"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	startTime := time.Now()
	// 1. receive jobs
	if len(os.Args) < 2 {
		log.Fatalf("wrong input, needs arguments <jobPath> <outputPath> and optional <pathToCfg>, e.x. singleTenant83.json outputFile.txt /home/franslukas/.kube/config")
	}

	var jobHandler jb.JobHandler

	var err error
	if len(os.Args) > 3 {
		jobHandler = jb.CreateJobHandler(os.Args[3])
	} else {
		jobHandler = jb.CreateJobHandler("")
	}

	helperFunctions.FatalErrCheck(err, "main: ")

	// 2. Parse to Job Class
	jobPath := os.Args[1]
	jobs, err := jb.ParseJson(jobPath, startTime)
	helperFunctions.FatalErrCheck(err, "main: ")


	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go initialTuning(job, jobHandler, &wg)
	}
	wg.Wait()

	println("train until convergence")
	trainUntilConvergence(jobHandler, jobs)

	storeTrainingData(jobs, os.Args[2])
}

func initialTuning(job *jb.Job, jobHandler jb.JobHandler, wg *sync.WaitGroup) {
	defer wg.Done()
	job.JobId = helperFunctions.GenerateId(constants.JOB_ID_LENGTH)
	println("testing reasonable batch size")
	batchSize := jobHandler.InitialTuning(job)

	//TODO not sure if this works fully (does it run epoch 1 over and over again?)
	*job.Epoch = 1
	for i := 0; i < 4; i++ {
		jobHandler.RunMiniEpoch(job, batchSize, i)
		*job.Epoch++
	}

	println("done with testing reasonable batch size")
}

func storeTrainingData(jobs []*jb.Job, outputFilePath string) {
	f, err := os.Create(outputFilePath)
	helperFunctions.FatalErrCheck(err, "storeTrainingData: ")

	defer f.Close()

	w := bufio.NewWriter(f)

	_, err = w.WriteString("[")
	helperFunctions.FatalErrCheck(err, "storeTrainingData: WriteString 1: ")

	for i, job := range jobs {
		output := job.GetTrainingData()
		if i != 0 {
			_, err = w.WriteString(",\n" + output)
		} else {
			_, err = w.WriteString("\n" + output)
		}
		helperFunctions.FatalErrCheck(err, "storeTrainingData: WriteString 2: ")
	}

	_, err = w.WriteString("\n]")
	helperFunctions.FatalErrCheck(err, "storeTrainingData: WriteString 3: ")

	err = w.Flush()
	helperFunctions.FatalErrCheck(err, "storeTrainingData: Flush: ")
}

func trainUntilConvergence(handler jb.JobHandler, jobs []*jb.Job) {
	allJobsDone := false

	functionsDeleted := make([]bool, len(jobs))

	for _, job := range jobs {
		*job.ActualTrainingStarted = true
	}

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

	budgetsForEpoch := make([]float64, len(jobs))
	for i, job := range jobs {
		// 4. Calculate number of functions we want to invoke
		budgetForeEpoch, err := job.CalculateBudgetForEpoch()

		helperFunctions.FatalErrCheck(err, "trainOneEpoch: ")

		budgetsForEpoch[i] = budgetForeEpoch
		fmt.Printf("\tbudget for job %s : %f\n", job.JobId, budgetForeEpoch)
	}

	// 5. Calculate number of functions we can invoke
	workerDeployment, serverDeployment := handler.GetDeploymentWithHighestMarginalUtility(jobs, budgetsForEpoch, outsideWorkers, outsideServers)

	for i, job := range jobs {
		//numberOfFunctionsToDeploy := handler.DeployableNumberOfFunctions(job, desiredNumberOfFunctions)
		numberOfFunctionsToDeploy := workerDeployment[i]
		fmt.Printf("actual number of workers: %d\n", numberOfFunctionsToDeploy)

		numberOfServersToDeploy := serverDeployment[i]
		fmt.Printf("actual number of servers: %d\n", numberOfServersToDeploy)

		job.SetNumberOfWorkers(numberOfFunctionsToDeploy)
		job.SetNumberOfServers(numberOfServersToDeploy)

		handler.DeleteExcessWorkers(job)
		handler.DeleteExcessParameterServers(job)
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


func executeTrainingOfOneEpoch(handler jb.JobHandler, job *jb.Job) {
	println("invoking functions")

	epochStartTime := time.Now()

	handler.InvokeFunctions(job)

	// print history events and loss estimation function
	//job.LeastSquaresTest()

	*job.Epoch++

	// update costs for functions
	cost := CostCalculator.CalculateCostForPods(job.JobId, handler.ClientSet, handler.MetricsClientSet, epochStartTime)
	job.UpdateFunctionCostsInHistory(cost)

	println("epoch is done")
}
