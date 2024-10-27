package config

import (
	"flag"
	"github.com/zasuchilas/gophermart/pkg/envflags"
)

var (
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	LogLevel             string
	EnvType              string
)

func ParseFlags() {
	flag.StringVar(&RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&DatabaseURI, "d", "", "database connection string")
	flag.StringVar(&AccrualSystemAddress, "r", "localhost:8081", "address of the accrual calculation service")
	flag.StringVar(&LogLevel, "l", "info", "logging level")
	flag.StringVar(&EnvType, "e", "production", "type of environment (production or develop)")
	flag.Parse()

	envflags.TryUseEnvString(&RunAddress, "RUN_ADDRESS")
	envflags.TryUseEnvString(&DatabaseURI, "DATABASE_URI")
	envflags.TryUseEnvString(&AccrualSystemAddress, "ACCRUAL_SYSTEM_ADDRESS")
	envflags.TryUseEnvString(&LogLevel, "LOG_LEVEL")
	envflags.TryUseEnvString(&EnvType, "ENV_TYPE")
}
