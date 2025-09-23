package util

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func GetenvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Errorf("environment variable %s=%q cannot be converted to an int", key, value))
	}
	return intValue
}

func GetenvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Errorf("environment variable %s=%q cannot be converted to a bool", key, value))
	}
	return boolValue
}

func GetenvDuration(key string, defaultValue time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	durationValue, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Errorf("environment variable %s=%q cannot be converted to a time.Duration", key, value))
	}
	return durationValue
}

func GetenvStr(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func MustGetenvInt(key string) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Errorf("environment variable %s must be set", key))
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Errorf("environment variable %s=%q cannot be converted to an int", key, value))
	}
	return intValue
}

func MustGetenvStr(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic(fmt.Errorf("environment variable %s must be set", key))
	}

	return value
}
