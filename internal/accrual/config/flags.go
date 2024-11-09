package config

import (
	"flag"
	"github.com/zasuchilas/gophermart/pkg/envflags"
	"time"
)

var (
	RunAddress      string
	DatabaseURI     string
	LogLevel        string
	EnvType         string
	WorkerPeriod    time.Duration
	WorkerPackLimit int
)

func ParseFlags() {
	flag.StringVar(&RunAddress, "a", "localhost:8081", "address and port to run server")
	flag.StringVar(&DatabaseURI, "d", "", "database connection string")
	flag.StringVar(&LogLevel, "l", "debug", "logging level")
	flag.StringVar(&EnvType, "e", "production", "type of environment (production or develop)")
	flag.DurationVar(&WorkerPeriod, "w", 3*time.Second, "calculate accrual worker period")
	flag.IntVar(&WorkerPackLimit, "p", 25, "calculate accrual worker pack limit")
	flag.Parse()

	envflags.TryUseEnvString(&RunAddress, "RUN_ADDRESS")
	envflags.TryUseEnvString(&DatabaseURI, "DATABASE_URI")
	envflags.TryUseEnvString(&LogLevel, "LOG_LEVEL")
	envflags.TryUseEnvString(&EnvType, "ENV_TYPE")
	envflags.TryUseEnvDuration(&WorkerPeriod, "WORKER_PERIOD")
	envflags.TryUseEnvInt(&WorkerPackLimit, "WORKER_PACK_LIMIT")
}
