package utils

import (
	"os"
	"sync"
)

const (
	envDevMode = "DEV_MODE"
)

var (
	devMode bool
	once    sync.Once
)

// DevMode enables devmode if the DEV_MODE env var is set to anything
func DevMode() bool {
	once.Do(func() {
		devMode = getDevMode()
	})
	return devMode
}

func getDevMode() bool {
	flag := os.Getenv(envDevMode)
	if flag == "" {
		return false
	} else {
		return true
	}
}
