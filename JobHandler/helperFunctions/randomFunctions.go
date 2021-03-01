package helperFunctions

import (
	"math/rand"
)


func GenerateId(idLength int) string {
	id := make([]byte, idLength)
	const charSet = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < idLength; i++ {
		id[i] = charSet[rand.Int() % idLength]
	}
	return string(id)
}
