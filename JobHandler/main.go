package main

import (
	"jobHandler/helperFunctions"
	jb "jobHandler/jobHandler"
	"log"
	"math/rand"
	"os"
	"time"
)

func testHyperbolaEstimation() {
	x := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	y := []float64{0.508112, 0.367166, 0.327031, 0.300430, 0.280054, 0.262924, 0.248206, 0.234580, 0.221567, 0.209484, 0.199290, 0.190342}

	fit := helperFunctions.HyperbolaLeastSquares(x, y)


	for i, f := range x {
		res := helperFunctions.EstimateYValueInHyperbola(f, fit)
		println("x: ", f, " y: ", y[i], "est: ", res)
	}
	println("done\n\n")

	//fit2 := helperFunctions.HyperbolaLeastSquares(y, x)


	for i, f := range y {
		res := helperFunctions.EstimateXValueInHyperbola(f, fit)
		println("y: ", f, " x: ", x[i], "est: ", res)
	}
	println("done\n\n")
}

func main() {
	testHyperbolaEstimation()
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
	*job.History = append(*job.History, jb.HistoryEvent{Loss: 1.0, Epoch: 1})
	*job.Epoch++
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


	println(&jobHandler)
	//
	//
	//job.JobId = helperFunctions.GenerateId(10)
	//
	//// 3. If done, store gradients and remove job from queue.
	////for !job.IsDone() {
	//println("train until convergence")
	//trainUntilConvergence(jobHandler, job)
}

func trainUntilConvergence(handler jb.JobHandler, job jb.Job) {

	for !job.IsDone() {
		// 4. Calculate number of functions we want to invoke
		desiredNumberOfFunctions := job.CalculateNumberOfFunctions()

		// 5. Calculate number of functions we can invoke
		numberOfFunctionsToDeploy := handler.DeployableNumberOfFunctions(job, desiredNumberOfFunctions)
		numberOfFunctionsToDeploy = 1
		println(numberOfFunctionsToDeploy)

		activeFunctions := (*handler.InstancesPerJob)[job.JobId]

		if activeFunctions < numberOfFunctionsToDeploy {
			handler.DeployFunctions(job, numberOfFunctionsToDeploy)
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

	// 7. Await response from all invoked functions (loss)
	println("waiting for invocation responses")
	//handler.AwaitResponse(job)

	// print history events and loss estimation function
	job.LeastSquaresTest()

	// 8. aggregate history, and repeat from step 3.
	handler.InvokeAggregator(job, numberOfFunctionsToInvoke)

	*job.Epoch++

	println("job is done")
}
