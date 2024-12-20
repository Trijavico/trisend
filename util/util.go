package util

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strconv"
)

func GetRandomID(size int) string {
	b := make([]byte, size)
	rand.Read(b)

	return base64.URLEncoding.EncodeToString(b)
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
