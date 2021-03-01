package main

import (
	"jobHandler/helperFunctions"
	"jobHandler/structs"
	"log"
	"os"
)

func main() {
	// 1. receive job
	if len(os.Args) < 2 {
		log.Fatalf("wrong input, needs argumetn jobPath")
	}

	// 2. Parse to Job Class
	jobPath := os.Args[1]
	job, err := structs.ParseJson(jobPath)
	helperFunctions.FatalErrCheck(err, "main: ")

	// 3. If done, store gradients and remove job from queue.
	if job.IsDone() {
		println("job is done")
	}

	// 4. Calculate number of functions we want to invoke


	// 5. Calculate number of functions we can invoke

	// 6. Invoke functions asynchronously

	// 7. Await response from all invoked functions (loss)

	// 8. Save history, and repeat from step 3.
}
