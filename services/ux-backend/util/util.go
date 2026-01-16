package util

import (
	"fmt"
	"os"
)

func ReadEnvVar[T any](envVarName string, defaultValue T, parser func(str string) (T, error)) (T, error) {
	str := os.Getenv(envVarName)
	if str == "" {
		return defaultValue, nil
	}

	value, err := parser(str)
	if err != nil {
		return *new(T), fmt.Errorf("malformed user-defined %s value %s: %v", envVarName, str, err)
	}
	return value, nil
}
