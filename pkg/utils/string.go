package utils

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var randSource = rand.NewSource(time.Now().UnixNano())
var randGenerator = rand.New(randSource)

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[randGenerator.Intn(len(charset))]
	}
	return string(b)
}
