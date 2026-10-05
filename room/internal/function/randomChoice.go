package function

import (
	"math/rand/v2"
)

func RandomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	randomIndex := rand.IntN(len(choices))
	return choices[randomIndex]
}
