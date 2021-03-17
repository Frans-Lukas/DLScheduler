package main

import (
	"fmt"
	"jobHandler/CostCalculator"
	"jobHandler/helperFunctions"
	jb "jobHandler/jobHandler"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	res := helperFunctions.Python3DParabolaLeastSquares([]float64{-1., -0.89473684, -0.78947368, -0.68421053, -0.57894737, -0.47368421, -0.36842105, -0.26315789, -0.15789474, -0.05263158, 0.05263158, 0.15789474, 0.26315789, 0.36842105, 0.47368421, 0.57894737, 0.68421053, 0.78947368, 0.89473684, 1.}, []float64{-1., -0.89473684, -0.78947368, -0.68421053, -0.57894737, -0.47368421, -0.36842105, -0.26315789, -0.15789474, -0.05263158, 0.05263158, 0.15789474, 0.26315789, 0.36842105, 0.47368421, 0.57894737, 0.68421053, 0.78947368, 0.89473684, 1.}, []float64{2.655, 2.09876731, 1.60901662, 1.18574792, 0.82896122, 0.53865651, 0.3148338, 0.15749307, 0.06663435, 0.04225762, 0.08436288, 0.19295014, 0.36801939, 0.60957064, 0.91760388, 1.29211911, 1.73311634, 2.24059557, 2.81455679, 3.455}, []float64{0, 0, 1, 2})
	println(helperFunctions.Python3DParabolaLeastSquaresEstimateH(-1, 1, res))


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
		deployment := handler.GetDeploymentWithHighestMarginalUtility(jobs, maxFuncs)

		//numberOfFunctionsToDeploy := handler.DeployableNumberOfFunctions(job, desiredNumberOfFunctions)
		numberOfFunctionsToDeploy := deployment[0]
		fmt.Printf("actual number of funcs: %d\n", numberOfFunctionsToDeploy)

		activeFunctions := (*handler.InstancesPerJob)[job.JobId]

		if activeFunctions < numberOfFunctionsToDeploy {
			handler.DeployFunctions(job)
			(*handler.InstancesPerJob)[job.JobId] = numberOfFunctionsToDeploy
		} else if activeFunctions > numberOfFunctionsToDeploy {
			numberOfVmsToKill := activeFunctions - numberOfFunctionsToDeploy
			//kill functions from numberOfFunctionsToDeploy to numberOfFunctionsToDeploy + activeFunctions
			startRange := numberOfFunctionsToDeploy
			endRange := numberOfFunctionsToDeploy + numberOfVmsToKill - 1
			handler.DeleteNuclioFunctionsInJob(job, int(startRange), int(endRange))
		}

		// TODO: wait until function is fully ready before invoking, sleep as a temp solution.
		err := handler.WaitForAllWorkerPods(job, "nuclio", time.Second*10)
		helperFunctions.FatalErrCheck(err, "waitForAllWorkerPods")

		trainOneEpoch(handler, job, numberOfFunctionsToDeploy)

		// TODO check if this works
		//handler.DeleteNuclioFunctionsInJob(job)
		//if we do not include epoch in pod name we will have to wait for them to delete
	}
}

func trainOneEpoch(handler jb.JobHandler, job jb.Job, numberOfFunctionsToInvoke uint) {
	println("invoking functions")
	handler.InvokeFunctions(job, int(numberOfFunctionsToInvoke))

	// print history events and loss estimation function
	job.LeastSquaresTest()

	*job.Epoch++

	// update costs for functions
	cost := CostCalculator.CalculateCostForPods(job.JobId, handler.ClientSet, handler.MetricsClientSet)
	job.UpdateAverageFunctionCost(cost)

	println("job is done")
}
