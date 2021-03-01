package structs

import (
	"encoding/json"
	"io/ioutil"
	"jobHandler/helperFunctions"
	"os"
)

type Job struct {
	Budget          float64 `json:"budget"`
	TargetLoss      float64 `json:"targetLoss"`
	ImageUrl        string  `json:"imageUrl"`
	CurrentCost     float64
	JobId           string
	FunctionIds     map[int]bool
	FunctionChannel chan int
	History         []HistoryEvent
}

func ParseJson(jsonPath string) (Job, error) {
	file, err := os.Open(jsonPath)

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
	return !job.historyIsEmpty() && job.History[len(job.History)-1].Loss <= job.TargetLoss
}

func (job Job) historyIsEmpty() bool {
	return len(job.History) == 0
}

func (job Job) CalculateNumberOfFunctions() uint {
	if job.historyIsEmpty() {
		return 1
	}
	return 5
}

func (job Job) FunctionsHaveFinished() bool {
	for _, functionIsDone := range job.FunctionIds {
		if functionIsDone == false {
			return false
		}
	}
	return true
}
