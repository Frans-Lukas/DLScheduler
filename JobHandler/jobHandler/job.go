package jobHandler

import (
	"encoding/json"
	"fmt"
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
	Epoch           *int
	FunctionChannel *chan int
	History         *[]HistoryEvent
}

func ParseJson(jsonPath string) (Job, error) {
	file, err := os.Open(jsonPath)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	byteValue, err := ioutil.ReadAll(file)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	var job Job

	history := make([]HistoryEvent, 0)
	epoch := 0
	tmpChan := make(chan int)
	job.FunctionChannel = &tmpChan
	job.FunctionIds = make(map[int]bool, 0)
	job.History = &history
	job.Epoch = &epoch

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
	budgetSurpassed := job.CurrentCost >= job.Budget
	if budgetSurpassed {
		println("budget surpassed for job: ", job.JobId)
	}
	return budgetSurpassed
}

func (job Job) lossReached() bool {
	lossReached := !job.historyIsEmpty() && (*job.History)[len(*job.History)-1].Loss <= job.TargetLoss
	if lossReached {
		println("loss reached for job: ", job.JobId)
	}
	return lossReached
}

func (job Job) historyIsEmpty() bool {
	return len(*job.History) == 0
}

func (job Job) CalculateNumberOfFunctions() uint {
	if job.historyIsEmpty() {
		return 1
	}
	return 5
}

func (job Job) FunctionsHaveFinished() bool {
	for _, functionIsDone := range job.FunctionIds {
		println(functionIsDone)
		if functionIsDone == false {
			return false
		}
	}
	return true
}

func (job Job) LeastSquaresTest() {
	println("History log:")
	x := make([]float64, 0)
	y := make([]float64, 0)
	for _, historyEvent := range *job.History {
		fmt.Printf("epoch: %d, loss: %f, accuracy; %f\n", historyEvent.Epoch, historyEvent.Loss, historyEvent.Accuracy)
		x = append(x, float64(historyEvent.Epoch))
		y = append(y, historyEvent.Loss)
	}
	println("performing estimation")

	function := helperFunctions.PolynomialLeastSquares(x, y)
	fmt.Printf("y = %f + %fx + %fx^2\n", function[0], function[1], function[2])
	for i := 0; i < 100; i++ {
		helperFunctions.PerformEstimationWithFunctions(function, float64(i))
	}
}

//TODO has not been checked if it works
func (job Job) MarginalUtilityCheck(numWorkers float64) float64 {
	println("History log:")
	x := make([]float64, 0)
	y := make([]float64, 0)
	for _, historyEvent := range *job.History {
		fmt.Printf("numFunctions: %d, loss: %f\n", historyEvent.NumWorkers, historyEvent.Time)
		x = append(x, float64(historyEvent.NumWorkers))
		y = append(y, historyEvent.Time)
	}
	println("performing estimation")

	function := helperFunctions.HyperbolaLeastSquares(x, y)
	fmt.Printf("y = %f + %fx", function[0], function[1])

	oldWorkers := numWorkers - 1
	oldTime := helperFunctions.EstimateValueInHyperbola(oldWorkers, function)

	newTime := helperFunctions.EstimateValueInHyperbola(numWorkers, function)

	return oldTime - newTime
}