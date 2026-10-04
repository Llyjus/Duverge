package function

import (
	"math/rand/v2"
	"strconv"
)

func GenerateRoomCode() string {
	n := rand.IntN(90000) + 10000
	return strconv.Itoa(n)
}
