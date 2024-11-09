package util

import (
	"math/rand"
	"time"
)

var fastRand = rand.New(rand.NewSource(time.Now().UnixNano()))
var charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func GetRandomID() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[fastRand.Intn(len(charset))]
	}
	return string(b)
}
