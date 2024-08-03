package utils

import (
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func GetOrDefault(name string, defaultv int) int {
	val, err := strconv.Atoi(os.Getenv(name))
	if err != nil || val == 0 {
		return defaultv
	}
	return val
}

var DEVICE_DISCOVERY_INTERVAL_SEC = time.Duration(GetOrDefault("DEVICE_DISCOVERY_INTERVAL_SEC", 30))
var ORCHESTRATOR_UPDATE_FREQUENCY_SECONDS = time.Duration(GetOrDefault("ORCHESTRATOR_UPDATE_FREQUENCY_SECONDS", 5))
var SESSION_TIMEOUT_CHECK_INTERVAL_SEC = time.Second * time.Duration(GetOrDefault("SESSION_TIMEOUT_CHECK_INTERVAL_SEC", 60000))
