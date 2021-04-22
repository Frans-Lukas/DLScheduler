package jobHandler

import (
	"encoding/json"
	"io/ioutil"
	"jobHandler/helperFunctions"
	"os"
)

type TestingErrors struct {
	errors map[string]float64
}

func ParseTestingErrorsFromJson(jsonPath string) *TestingErrors {
	var testingErrors TestingErrors

	file, err := os.Open(jsonPath)

	helperFunctions.FatalErrCheck(err, "ParseTestingErrorsFromJson: ")

	byteValue, err := ioutil.ReadAll(file)

	helperFunctions.FatalErrCheck(err, "ParseTestingErrorsFromJson: ")

	err = json.Unmarshal(byteValue, &testingErrors.errors)

	helperFunctions.FatalErrCheck(err, "ParseTestingErrorsFromJson: ")

	return &testingErrors
}

func (testingErrors TestingErrors) ApplyError(originalValue float64, errorName string) float64 {
	errorFloat := testingErrors.errors[errorName]

	if errorFloat != 0 {
		return originalValue * errorFloat
	} else {
		return originalValue
	}
}

func (testingErrors TestingErrors) GetError(errorName string) float64 {
	errorFloat := testingErrors.errors[errorName]
	return errorFloat
}