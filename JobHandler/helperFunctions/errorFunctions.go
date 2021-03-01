package helperFunctions

import "log"

func FatalErrCheck(err error, s string) {
	if err != nil {
		log.Fatalf(s, err.Error())
	}
}
