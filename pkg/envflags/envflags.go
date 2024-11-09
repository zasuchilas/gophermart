package envflags

import (
	"log"
	"os"
	"strconv"
	"time"
)

func TryUseEnvString(flagValue *string, envName string) {
	if env := os.Getenv(envName); env != "" {
		*flagValue = env
	}
}

func TryUseEnvDuration(flagValue *time.Duration, envName string) {
	if env := os.Getenv(envName); env != "" {
		v, err := time.ParseDuration(env)
		if err != nil {
			log.Panicf("error trying parse ENV %s (TryUseEnvDuration, error: %s)", envName, err.Error())
		}
		*flagValue = v
	}
}

func TryUseEnvInt(flagValue *int, envName string) {
	if env := os.Getenv(envName); env != "" {
		v, err := strconv.Atoi(env)
		if err != nil {
			log.Panicf("error trying parse ENV %s (TryUseEnvInt, error: %s)", envName, err.Error())
		}
		*flagValue = v
	}
}
