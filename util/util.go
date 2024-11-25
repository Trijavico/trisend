package util

import (
	"math/rand"
	"os"
	"strconv"
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

func GetEnvStr(name string, callback string) string {
	value := os.Getenv(name)
	if value == "" {
		return callback
	}

	return value
}

func GetEnvInt(name string, callback int) int {
	value := os.Getenv(name)
	if value == "" {
		return callback
	}

	converted, err := strconv.Atoi(value)
	if err != nil {
		return callback
	}

	return converted
}
