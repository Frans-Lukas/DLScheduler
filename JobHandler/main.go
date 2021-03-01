package main

import (
	"encoding/json"
	"io/ioutil"
	"jobHandler/structs"
	"log"
	"os"
)

func main() {
	// 1. receive job
	if len(os.Args) < 2 {
		log.Fatalf("wrong input, needs argumetn jobPath")
	}

	jobPath := os.Args[1]
	file, err  := os.Open(jobPath)

	fatalErrCheck(err, "main: ")

	byteValue, err := ioutil.ReadAll(file)

	fatalErrCheck(err, "main: ")

	var job structs.Job

	err = json.Unmarshal(byteValue, &job)

	fatalErrCheck(err, "main: ")

	println(job.Budget)
	println(job.TargetLoss)
	println(job.ImageUrl)

	// 2. Parse to Job Class

	// 3. If done, store gradients and remove job from queue.

	// 4. Calculate number of functions we want to invoke

	// 5. Calculate number of functions we can invoke

	// 6. Invoke functions asynchronously

	// 7. Await response from all invoked functions (loss)

	// 8. Save history, and repeat from step 3.
}

func fatalErrCheck(err error, s string) {
	if err != nil {
		log.Fatalf(s, err.Error())
	}
}
