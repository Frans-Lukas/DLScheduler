package jobHandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jobHandler/helperFunctions"
	"math"
	"os"
)

type Job struct {
	Budget                float64 `json:"budget"`
	TargetLoss            float64 `json:"targetLoss"`
	ImageUrl              string  `json:"imageUrl"`
	DataSetSize           int     `json:"dataSetSize"`
	CurrentCost           float64
	JobId                 string
	PodNames              map[string]bool
	Epoch                 *int
	FunctionChannel       *chan string
	AverageFunctionCost   float64
	NumberOfFunctionsUsed uint
	NumberOfWorkers       int
	NumberOfServers       int
	SchedulerIp           *string
	History               *[]HistoryEvent
	MarginalUtilityFunc   *[]float64
}

func ParseJson(jsonPath string) (Job, error) {
	file, err := os.Open(jsonPath)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	byteValue, err := ioutil.ReadAll(file)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	var job Job

	ipString := ""

	history := make([]HistoryEvent, 1)
	marginalUtilityFunc := make([]float64, 1)
	history[0] = HistoryEvent{Epoch: 1, Loss: 1.0}
	epoch := 2
	tmpChan := make(chan string)
	job.FunctionChannel = &tmpChan
	job.PodNames = make(map[string]bool, 0)
	job.History = &history
	job.Epoch = &epoch
	job.NumberOfWorkers = 1
	job.NumberOfServers = 1
	job.SchedulerIp = &ipString
	job.MarginalUtilityFunc = &marginalUtilityFunc

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
	// always contains (loss = 1, epoch = 1)
	return len(*job.History) <= 1
}

func (job Job) CalculateNumberOfFunctions() uint {
	if job.historyIsEmpty() {
		return 1
	}

	epochsTillConvergence := job.CalculateEpochsTillConvergence()
	fmt.Printf("epochs until convergence: %d\n", epochsTillConvergence)

	maxFunctions := job.maxFunctionsWithRemainingBudget()
	fmt.Printf("maxFunctions: %d\n", maxFunctions)

	functions := job.functionsForNextEpoch(maxFunctions, epochsTillConvergence)
	fmt.Printf("functions: %d\n", functions)

	return functions
}

func (job Job) FunctionsHaveFinished() bool {
	for _, functionIsDone := range job.PodNames {
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

	function := helperFunctions.HyperbolaLeastSquares(x, y)
	fmt.Printf("y = %f + %fx\n", function[0], function[1])
	for i := 1; i <= 100; i++ {
		fmt.Printf("%f\n", helperFunctions.EstimateYValueInHyperbolaFunction(float64(i), function))
	}
}

func (job Job)UpdateMarginalUtilityFunc() {
	if job.historyIsEmpty() {
		return //TODO find better solution for this
	}

	x := make([]float64, 0)
	y := make([]float64, 0)
	h := make([]float64, 0)
	for _, historyEvent := range *job.History {
		fmt.Printf("numFunctions: %d, steps/s: %f\n", historyEvent.NumWorkers, 1/historyEvent.Time)
		x = append(x, float64(historyEvent.NumWorkers))
		y = append(y, float64(historyEvent.NumServers))
		h = append(h, historyEvent.Time)
	}

	previousEstimation := *job.MarginalUtilityFunc
	if len(previousEstimation) < 4 {
		previousEstimation = []float64{0, 0, 1, 2}
	}

	*job.MarginalUtilityFunc = helperFunctions.Python3DParabolaLeastSquares(x, y, h, previousEstimation) //TODO check if this should be done with polynomial least squares and steps/s instead of time (check optimus)
	fmt.Printf("y = %f + %fx", (*job.MarginalUtilityFunc)[0], (*job.MarginalUtilityFunc)[1])
}

//TODO has not been checked if it works
func (job Job) MarginalUtilityCheck(numWorkers uint, numServers uint, oldWorkers uint, oldServers uint, maxWorkers uint) float64 {
	if numWorkers > maxWorkers {
		return -1
	}

	if job.historyIsEmpty() {
		return 1 //TODO find better solution for this
	}

	if len(*job.MarginalUtilityFunc) == 0 {
		job.UpdateMarginalUtilityFunc()
	}
	oldStepsPerSec := helperFunctions.Python3DParabolaLeastSquaresEstimateH(float64(oldWorkers), float64(oldServers), *job.MarginalUtilityFunc)

	newStepsPerSec := helperFunctions.Python3DParabolaLeastSquaresEstimateH(float64(numWorkers), float64(numServers), *job.MarginalUtilityFunc)

	return newStepsPerSec - oldStepsPerSec
}

func (job Job) CalculateEpochsTillConvergence() uint {
	x := make([]float64, 0)
	y := make([]float64, 0)
	for _, historyEvent := range *job.History {
		x = append(x, float64(historyEvent.Epoch))
		y = append(y, historyEvent.Loss)
	}

	function := helperFunctions.HyperbolaLeastSquares(x, y)

	convergenceEpoch := helperFunctions.EstimateXValueInHyperbolaFunction(job.TargetLoss, function)

	return uint(int(math.Ceil(convergenceEpoch)) - *job.Epoch) //TODO should we have some sort of "optimism" deterrent (ex. multiply by 1.1)
}

func (job Job) maxFunctionsWithRemainingBudget() uint {
	currentBudget := job.Budget - job.CurrentCost

	return uint(currentBudget / job.costPerFunction())
}

func (job Job) costPerFunction() float64 {
	if job.AverageFunctionCost == 0 {
		println("job does not have an average function cost yet")
		return 10
	} else {
		return job.AverageFunctionCost
	}
}

func (job Job) functionsForNextEpoch(functions uint, epochs uint) uint {
	suggestedNumber := float64(functions / epochs)
	return uint(math.Max(suggestedNumber, 1)) //TODO make this take into account that fewer functions are used for later epochs
}

func (job Job) UpdateAverageFunctionCost(cost float64) {
	job.AverageFunctionCost = ((job.AverageFunctionCost * float64(job.NumberOfFunctionsUsed)) + cost) / float64(job.NumberOfFunctionsUsed+(*job.History)[len(*job.History) - 1].NumWorkers)
}
