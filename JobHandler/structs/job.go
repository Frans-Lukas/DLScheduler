package structs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"jobHandler/helperFunctions"
)

type Job struct {
	Budget     	float64 `json:"budget"`
	TargetLoss 	float64 `json:"targetLoss"`
	ImageUrl   	string `json:"imageUrl"`
	CurrentCost float64
	History     []HistoryEvent
}

func ParseJson(jsonPath string) (Job, error) {
	file, err  := os.Open(jsonPath)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	byteValue, err := ioutil.ReadAll(file)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	var job Job

	err = json.Unmarshal(byteValue, &job)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	println(job.Budget)
	println(job.TargetLoss)
	println(job.ImageUrl)

	return job, nil
}

func (job Job) IsDone() bool {
	return job.lossReached() || job.budgetSurpassed()
}

func (job Job) budgetSurpassed() bool {
	return job.CurrentCost >= job.Budget
}

func (job Job) lossReached() bool {
	return len(job.History) != 0 && job.History[len(job.History)-1].Loss <= job.TargetLoss
}