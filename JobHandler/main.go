package main

import (
	"jobHandler/helperFunctions"
	"jobHandler/jobHandler"
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

	var handler jobHandler.JobHandler

	var err error
	if len(os.Args) > 2 {
		handler = jobHandler.CreateJobHandler(os.Args[2])
	} else {
		handler = jobHandler.CreateJobHandler("")
	}

	helperFunctions.FatalErrCheck(err, "main: ")

	// 2. Parse to Job Class
	jobPath := os.Args[1]
	job, err := jobHandler.ParseJson(jobPath)
	helperFunctions.FatalErrCheck(err, "main: ")

	job.JobId = helperFunctions.GenerateId(10)

	// 3. If done, store gradients and remove job from queue.
	//for !job.IsDone() {

	// 4. Calculate number of functions we want to invoke
	desiredNumberOfFunctions := job.CalculateNumberOfFunctions()

	// 5. Calculate number of functions we can invoke
	numberOfFunctionsToDeploy := handler.DeployableNumberOfFunctions(job, desiredNumberOfFunctions)
	println(numberOfFunctionsToDeploy)

	// 6. Invoke functions asynchronously
	handler.DeployFunctions(job, numberOfFunctionsToDeploy)

	// TODO: wait until function is fully ready before invoking, sleep as a temp solution.
	time.Sleep(time.Second * 4)

	trainOneEpoch(handler, job, numberOfFunctionsToDeploy)
	//
	//}
}

func trainOneEpoch(handler jobHandler.JobHandler, job jobHandler.Job, numberOfFunctionsToInvoke uint) {
	println("invoking functions")
	handler.InvokeFunctions(job, int(numberOfFunctionsToInvoke))

	// 7. Await response from all invoked functions (loss)
	println("waiting for invocation responses")
	handler.AwaitResponse(job)

	// 8. aggregate history, and repeat from step 3.
	handler.InvokeAggregator(job, numberOfFunctionsToInvoke)

	job.Epoch++
	println("job is done")
}
