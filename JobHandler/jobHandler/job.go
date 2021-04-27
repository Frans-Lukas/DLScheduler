package jobHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"jobHandler/helperFunctions"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Job struct {
	Budget                float64 `json:"budget"`
	TargetLoss            float64 `json:"targetLoss"`
	ImageUrl              string  `json:"imageUrl"`
	DataSetSize           int     `json:"dataSetSize"`
	ScriptPath            string  `json:"scriptPath"`
	TestingErrorsPath     string  `json:"testingErrorsPath"`
	testingErrors         *TestingErrors
	CurrentCost           float64
	JobId                 string
	PodNames              map[string]bool
	DeployedPods          *[]string
	Epoch                 *int
	FunctionChannel       *chan string
	NumberOfFunctionsUsed uint
	NumberOfParts         int
	SchedulerIp           *string
	History               *[]HistoryEvent
	MarginalUtilityFunc   *[]float64
	InitialTuning         *bool
	workersMutex          sync.Mutex
	NumberOfWorkers       uint
	serversMutex          sync.Mutex
	NumberOfServers       uint
	isTraining            bool
	isTrainingMutex       sync.Mutex
	CostFunc              *[]float64
}

func ParseJson(jsonPath string) ([]*Job, error) {
	file, err := os.Open(jsonPath)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	byteValue, err := ioutil.ReadAll(file)

	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	var jobs []*Job

	err = json.Unmarshal(byteValue, &jobs)
	helperFunctions.FatalErrCheck(err, "ParseJson: ")

	for _, job := range jobs {
		ipString := ""

		history := make([]HistoryEvent, 0)
		marginalUtilityFunc := make([]float64, 0)
		costFunc := make([]float64, 0)
		deployedPods := make([]string, 0)
		//history[0] = HistoryEvent{Epoch: 1, Loss: 1.0}
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
		job.CostFunc = &costFunc
		job.isTraining = false
		tmpInitialTuning := false
		job.InitialTuning = &tmpInitialTuning
		job.testingErrors = ParseTestingErrorsFromJson(job.TestingErrorsPath)
		job.DeployedPods =  &deployedPods

		println(job.Budget)
		println(job.TargetLoss)
		println(job.ImageUrl)

		job.testingErrors.ApplyError(1, "marginalUtility2")
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

func (job *Job) CalculateBudgetForEpoch() (float64, error) {
	println("CalculateBudgetForEpoch: ")
	if job.historyIsEmpty() {
		return -1, errors.New("CalculateBudgetForEpoch: empty history")
	}

	epochsTillConvergence := job.CalculateEpochsTillConvergence()
	fmt.Printf("\tepochs until convergence: %d\n", epochsTillConvergence)

	budgetForEpoch := job.budgetForNextEpoch(epochsTillConvergence)
	fmt.Printf("\tbudgetForEpoch: %d\n", budgetForEpoch)

	return budgetForEpoch, nil
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

func (job *Job) UpdateCostFunc() {
	if job.historyIsEmpty() {
		return //TODO find better solution for this
	}

	x := make([]float64, 0)
	y := make([]float64, 0)
	h := make([]float64, 0)
	for _, historyEvent := range *job.History {
		if historyEvent.Cost == -1 {
			println("event with unset cost!")
		} else {
			//fmt.Printf("numWorkers: %d, numServers: %d, cost: %f\n", historyEvent.NumWorkers, historyEvent.NumServers, 1/historyEvent.Cost)
			x = append(x, float64(historyEvent.NumWorkers))
			y = append(y, float64(historyEvent.NumServers))
			h = append(h, historyEvent.Cost)
		}
	}

	previousEstimation := *job.CostFunc
	if len(previousEstimation) < 5 {
		previousEstimation = []float64{1, 1, 1, 1, 1}
	}

	*job.CostFunc = helperFunctions.Python3DParabolaLeastSquares(x, y, h, previousEstimation, "costEstimation")
}

func (job *Job) UpdateMarginalUtilityFunc() {
	if job.historyIsEmpty() {
		return //TODO find better solution for this
	}

	x := make([]float64, 0)
	y := make([]float64, 0)
	h := make([]float64, 0)
	for _, historyEvent := range *job.History {
		//fmt.Printf("numWorkers: %d, numServers: %d, steps/s: %f\n", historyEvent.NumWorkers, historyEvent.NumServers, 1/historyEvent.Time)
		x = append(x, float64(historyEvent.NumWorkers))
		y = append(y, float64(historyEvent.NumServers))
		h = append(h, 1/historyEvent.Time) // conversion to steps/s
	}

	previousEstimation := *job.MarginalUtilityFunc
	if len(previousEstimation) < 5 {
		previousEstimation = []float64{1, 1, 1, 1, 1}
	}

	//TODO check if this should be done with polynomial least squares and steps/s instead of time (check optimus)
	*job.MarginalUtilityFunc = helperFunctions.Python3DParabolaLeastSquares(x, y, h, previousEstimation, "marginalUtil")
}

//TODO has not been checked if it works
func (job *Job) MarginalUtilityCheck(numWorkers uint, numServers uint, oldWorkers uint, oldServers uint, budget float64) float64 {
	if job.testingErrors.GetError("staticWorkers") >= 1 {
		if numWorkers != uint(job.testingErrors.GetError("staticWorkers")) {
			println("\tERROR: not matching exceed static workers", job.testingErrors.GetError("staticWorkers"))
			return -1
		}
	}

	if job.testingErrors.GetError("staticServers") >= 1 {
		if numServers != uint(job.testingErrors.GetError("staticServers")) {
			println("\tERROR: not matching static servers", job.testingErrors.GetError("staticServers"))
			return -1
		}
	}

	if budget <= 0 {
		println("\tERROR: Would go over budget.")
		return -1
	}

	if job.testingErrors.GetError("maxWorkers") >= 1 {
		if float64(numWorkers) > job.testingErrors.GetError("maxWorkers") {
			println("\tERROR: Would exceed maxWorkers", job.testingErrors.GetError("maxServers"))
			return -1
		}
	}

	if job.testingErrors.GetError("maxServers") >= 1 {
		if float64(numServers) > job.testingErrors.GetError("maxServers") {
			println("\tERROR: Would exceed maxServers", job.testingErrors.GetError("maxServers"))
			return -1
		}
	}


	if job.historyIsEmpty() {
		return 1 //TODO find better solution for this
	}

	if len(*job.MarginalUtilityFunc) == 0 {
		job.UpdateMarginalUtilityFunc()
	}

	if len(*job.CostFunc) == 0 {
		job.UpdateCostFunc()
	}

	cost, err := helperFunctions.Python3DPolynomialEstimateH(float64(numWorkers), float64(numServers), *job.CostFunc)
	helperFunctions.FatalErrCheck(err, "MarginalUtilityCheck: ")
	cost = job.testingErrors.ApplyError(cost, "costEstimation")

	if cost > budget {
		println("\tERROR: cost ", cost, " would exceed budget ", budget)
		return -1
	}

	oldStepsPerSec := -1.0

	if oldWorkers == 0 || oldServers == 0 {
		oldStepsPerSec = 0.0
	} else {
		oldStepsPerSec, err = helperFunctions.Python3DParabolaLeastSquaresEstimateH(float64(oldWorkers), float64(oldServers), *job.MarginalUtilityFunc)
		helperFunctions.FatalErrCheck(err, "MarginalUtilityCheck: ")
		oldStepsPerSec = job.testingErrors.ApplyError(oldStepsPerSec, "marginalUtility")
	}

	newStepsPerSec, err := helperFunctions.Python3DParabolaLeastSquaresEstimateH(float64(numWorkers), float64(numServers), *job.MarginalUtilityFunc)
	newStepsPerSec = job.testingErrors.ApplyError(newStepsPerSec, "marginalUtility")
	helperFunctions.FatalErrCheck(err, "MarginalUtilityCheck: ")

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
	convergenceEpoch = job.testingErrors.ApplyError(convergenceEpoch, "convergenceEpoch")

	remainingEpochs := int(math.Ceil(convergenceEpoch)) - (*job.Epoch - 1)

	if remainingEpochs < 0 {
		return 1
	} else {
		return uint(remainingEpochs)
	} //TODO should we have some sort of "optimism" deterrent (ex. multiply by 1.1)
}

func (job *Job) budgetForNextEpoch(epochs uint) float64 {
	currentBudget := job.Budget - job.CurrentCost

	if epochs == 0 {
		return currentBudget
	}

	suggestedBudget := currentBudget / float64(epochs)
	return suggestedBudget //TODO make this take into account that fewer functions are used for later epochs
}

func (job *Job) UpdateFunctionCostsInHistory(cost float64) {
	if len(*job.History) == 0 {
		return
	}

	for i, event := range *job.History {
		if event.Cost == -1 {
			(*job.History)[i].Cost = cost
		}
	}
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

	job.NumberOfServers = numberOfServers
}

func (job *Job) GetNumberOfServers() uint {
	job.serversMutex.Lock()
	defer job.serversMutex.Unlock()

	res := job.NumberOfServers
	return res
}

func (job *Job) SetNumberOfWorkers(numberOfWorkers uint) {
	job.workersMutex.Lock()
	defer job.workersMutex.Unlock()

	job.NumberOfWorkers = numberOfWorkers
}

func (job *Job) GetNumberOfWorkers() uint {
	job.workersMutex.Lock()
	defer job.workersMutex.Unlock()

	res := job.NumberOfWorkers
	return res
}

func (job *Job) GetTrainingData() string {
	var b strings.Builder

	b.WriteString("\t{\n")

	b.WriteString("\t\t\"id\": \"" + job.JobId + "\",\n")
	b.WriteString("\t\t\"epochs\":" + strconv.Itoa(*job.Epoch) + ",\n")
	b.WriteString("\t\t\"loss\":" + job.getLossHistoryString() + ",\n")
	b.WriteString("\t\t\"servers\":" + job.getServerHistoryString() + ",\n")
	b.WriteString("\t\t\"workers\":" + job.getWorkerHistoryString() + ",\n")
	b.WriteString("\t\t\"cost\":" + job.getCostHistoryString() + ",\n")
	b.WriteString("\t\t\"time\":" + job.getTimeHistoryString() + "\n")

	b.WriteString("\t}")

	return b.String()
}

func (job *Job) getLossHistoryString() string {
	var b strings.Builder

	b.WriteString("[ ")

	currEpoch := -1

	for i, event := range *job.History {
		if event.Epoch > currEpoch {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%f", event.Loss))

			currEpoch = event.Epoch
		}
	}

	b.WriteString(" ]")

	return b.String()
}

func (job *Job) getServerHistoryString() string {
	var b strings.Builder

	b.WriteString("[ ")

	currEpoch := -1

	for i, event := range *job.History {
		if event.Epoch > currEpoch {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(strconv.Itoa(int(event.NumServers)))

			currEpoch = event.Epoch
		}
	}

	b.WriteString(" ]")

	return b.String()
}

func (job *Job) getWorkerHistoryString() string {
	var b strings.Builder

	b.WriteString("[ ")

	currEpoch := -1

	for i, event := range *job.History {
		if event.Epoch > currEpoch {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(strconv.Itoa(int(event.NumWorkers)))

			currEpoch = event.Epoch
		}
	}

	b.WriteString(" ]")

	return b.String()
}

func (job *Job) getCostHistoryString() string {
	var b strings.Builder

	b.WriteString("[ ")

	currEpoch := -1

	for i, event := range *job.History {
		if event.Epoch > currEpoch {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%f", event.Cost))

			currEpoch = event.Epoch
		}
	}

	b.WriteString(" ]")

	return b.String()
}

func (job *Job) getTimeHistoryString() string {
	var b strings.Builder

	b.WriteString("[ ")

	currEpoch := -1

	for i, event := range *job.History {
		if event.Epoch > currEpoch {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%f", event.Time))

			currEpoch = event.Epoch
		}
	}

	b.WriteString(" ]")

	return b.String()
}