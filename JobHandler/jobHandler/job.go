package jobHandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"jobHandler/helperFunctions"
	"math"
	"os"
	"sync"
)

type Job struct {
	Budget                float64 `json:"budget"`
	TargetLoss            float64 `json:"targetLoss"`
	ImageUrl              string  `json:"imageUrl"`
	DataSetSize           int     `json:"dataSetSize"`
	ScriptPath            string  `json:"scriptPath"`
	CurrentCost           float64
	JobId                 string
	PodNames              map[string]bool
	Epoch                 *int
	FunctionChannel       *chan string
	AverageFunctionCost   float64
	NumberOfFunctionsUsed uint
	workersMutex          sync.Mutex
	numberOfWorkers       uint
	serversMutex          sync.Mutex
	numberOfServers       uint
	SchedulerIp           *string
	History               *[]HistoryEvent
	MarginalUtilityFunc   *[]float64
	isTraining            bool
	isTrainingMutex       sync.Mutex
}

func ParseJson(jsonPath string) ([]*Job, error) {
	file, err := os.Open(jsonPath)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	byteValue, err := ioutil.ReadAll(file)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	var jobs []*Job

	err = json.Unmarshal(byteValue, &jobs)

	for _, job := range jobs {
		ipString := ""

		history := make([]HistoryEvent, 0)
		marginalUtilityFunc := make([]float64, 1)
		//history[0] = HistoryEvent{Epoch: 1, Loss: 1.0}
		epoch := 2
		tmpChan := make(chan string)
		job.FunctionChannel = &tmpChan
		job.PodNames = make(map[string]bool, 0)
		job.History = &history
		job.Epoch = &epoch
		job.numberOfWorkers = 1
		job.numberOfServers = 1
		job.SchedulerIp = &ipString
		job.MarginalUtilityFunc = &marginalUtilityFunc
		job.isTraining = false

		helperFunctions.FatalErrCheck(err, "ParseJson: ")

		println(job.Budget)
		println(job.TargetLoss)
		println(job.ImageUrl)
	}

	return jobs, nil
}

func (job *Job) IsDone() bool {
	return job.lossReached() || job.budgetSurpassed()
}

func (job *Job) budgetSurpassed() bool {
	budgetSurpassed := job.CurrentCost >= job.Budget
	if budgetSurpassed {
		println("budget surpassed for job: ", job.JobId)
	}
	return budgetSurpassed
}

func (job *Job) lossReached() bool {
	lossReached := !job.historyIsEmpty() && (*job.History)[len(*job.History)-1].Loss <= job.TargetLoss
	if lossReached {
		println("target loss: ", job.TargetLoss)
		println("latest loss: ", (*job.History)[len(*job.History)-1].Loss)
		println("loss reached for job: ", job.JobId)
	}
	return lossReached
}

func (job *Job) historyIsEmpty() bool {
	// always contains (loss = 1, epoch = 1)
	return len(*job.History) <= 1
}

func (job *Job) CalculateNumberOfFunctions() uint {
	if job.historyIsEmpty() {
		return 2
	}

	epochsTillConvergence := job.CalculateEpochsTillConvergence()
	fmt.Printf("epochs until convergence: %d\n", epochsTillConvergence)

	maxFunctions := job.maxFunctionsWithRemainingBudget()
	fmt.Printf("maxFunctions: %d\n", maxFunctions)

	functions := job.functionsForNextEpoch(maxFunctions, epochsTillConvergence)
	fmt.Printf("functions: %d\n", functions)

	if functions < 2 {
		return 2
	}
	return functions
}

func (job *Job) FunctionsHaveFinished() bool {
	for _, functionIsDone := range job.PodNames {
		println(functionIsDone)
		if functionIsDone == false {
			return false
		}
	}
	return true
}

func (job *Job) LeastSquaresTest() {
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

func (job *Job) UpdateMarginalUtilityFunc() {
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
	if len(previousEstimation) < 5 {
		previousEstimation = []float64{0, 0, 1, 2, 0}
	}

	*job.MarginalUtilityFunc = helperFunctions.Python3DParabolaLeastSquares(x, y, h, previousEstimation, "marginalUtil") //TODO check if this should be done with polynomial least squares and steps/s instead of time (check optimus)
	fmt.Printf("y = %f + %fx", (*job.MarginalUtilityFunc)[0], (*job.MarginalUtilityFunc)[1])
}

//TODO has not been checked if it works
func (job *Job) MarginalUtilityCheck(numWorkers uint, numServers uint, oldWorkers uint, oldServers uint, maxFunctions uint) float64 {
	if numWorkers + numServers > maxFunctions {
		return -1
	}

	if job.historyIsEmpty() {
		return 1 //TODO find better solution for this
	}

	if len(*job.MarginalUtilityFunc) == 0 {
		job.UpdateMarginalUtilityFunc()
	}
	oldStepsPerSec, err := helperFunctions.Python3DParabolaLeastSquaresEstimateH(float64(oldWorkers), float64(oldServers), *job.MarginalUtilityFunc)

	helperFunctions.FatalErrCheck(err, "MarginalUtilityCheck")

	newStepsPerSec, err := helperFunctions.Python3DParabolaLeastSquaresEstimateH(float64(numWorkers), float64(numServers), *job.MarginalUtilityFunc)

	helperFunctions.FatalErrCheck(err, "MarginalUtilityCheck")

	return newStepsPerSec - oldStepsPerSec
}

func (job *Job) CalculateEpochsTillConvergence() uint {
	x := make([]float64, 0)
	y := make([]float64, 0)
	for _, historyEvent := range *job.History {
		x = append(x, float64(historyEvent.Epoch))
		y = append(y, historyEvent.Loss)
	}

	startingGuess := []float64{1, 1, 1}

	function := helperFunctions.Python3DParabolaLeastSquares(x, y, make([]float64, 1), startingGuess, "convergence")

	convergenceEpoch, err := helperFunctions.PythonParabolicLeastSquaresEstimateX(job.TargetLoss, function)

	helperFunctions.FatalErrCheck(err, "CalculateEpochsTillConvergence: ")

	remainingEpochs := int(math.Ceil(convergenceEpoch)) - (*job.Epoch - 1)

	if remainingEpochs < 0 {
		return 1
	} else {
		return uint(remainingEpochs)
	} //TODO should we have some sort of "optimism" deterrent (ex. multiply by 1.1)
}

func (job *Job) maxFunctionsWithRemainingBudget() uint {
	currentBudget := job.Budget - job.CurrentCost

	return uint(currentBudget / job.costPerFunction())
}

func (job *Job) costPerFunction() float64 {
	if job.AverageFunctionCost == 0 {
		println("job does not have an average function cost yet")
		return 10
	} else {
		return job.AverageFunctionCost
	}
}

func (job *Job) functionsForNextEpoch(functions uint, epochs uint) uint {
	if epochs == 0 {
		return functions
	}
	suggestedNumber := float64(functions / epochs)
	return uint(math.Max(suggestedNumber, 1)) //TODO make this take into account that fewer functions are used for later epochs
}

func (job *Job) UpdateAverageFunctionCost(cost float64) {
	if len(*job.History) == 0 {
		return
	}

	job.AverageFunctionCost = ((job.AverageFunctionCost * float64(job.NumberOfFunctionsUsed)) + cost) / float64(job.NumberOfFunctionsUsed+(*job.History)[len(*job.History)-1].NumWorkers)
}

func (job *Job) UpdateIsTraining(isTraining bool) {
	job.isTrainingMutex.Lock()
	defer job.isTrainingMutex.Unlock()

	job.isTraining = isTraining
}

func (job *Job) CheckIsTraining() bool {
	job.isTrainingMutex.Lock()
	defer job.isTrainingMutex.Unlock()

	res := job.isTraining
	return res
}

func (job *Job) SetNumberOfServers(numberOfServers uint) {
	job.serversMutex.Lock()
	defer job.serversMutex.Unlock()

	job.numberOfServers = numberOfServers
}

func (job *Job) GetNumberOfServers() uint {
	job.serversMutex.Lock()
	defer job.serversMutex.Unlock()

	res := job.numberOfServers
	return res
}

func (job *Job) SetNumberOfWorkers(numberOfWorkers uint) {
	job.workersMutex.Lock()
	defer job.workersMutex.Unlock()

	job.numberOfWorkers = numberOfWorkers
}

func (job *Job) GetNumberOfWorkers() uint {
	job.workersMutex.Lock()
	defer job.workersMutex.Unlock()

	res := job.numberOfWorkers
	return res
}