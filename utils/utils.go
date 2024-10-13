package utils

import (
	"math/rand"
	"strings"
)

var alphanumerics = [...]rune{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
	'J', 'K', 'L', 'M', 'N', 'O', 'P', 'q', 'r',
	's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '1',
	'2', '3', '4', '5', '6', '7', '8', '9', '0',
}

func GenerateRandomString(length uint) string {
	var key strings.Builder
	for range length {
		key.WriteRune(alphanumerics[rand.Intn(len(alphanumerics))])
	}

	return key.String()
}
