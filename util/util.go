package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetRandomID(size int) string {
	b := make([]byte, size)
	rand.Read(b)

	return base64.RawURLEncoding.EncodeToString(b)
}

func GetFingerPrint(key string) (string, error) {
	splitted := strings.Split(key, " ")
	if len(splitted) < 2 {
		return "", fmt.Errorf("invalid SSH public key format")
	}

	cleanStr := strings.ReplaceAll(splitted[1], "\n", "")
	sshBytes, err := base64.StdEncoding.DecodeString(cleanStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode ssh key: %v", err)
	}

	shaHash := sha256.Sum256(sshBytes)
	fingerprint := base64.RawStdEncoding.EncodeToString(shaHash[:])

	return fingerprint, nil
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
