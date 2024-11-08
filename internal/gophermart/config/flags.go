package config

import (
	"flag"
	"github.com/zasuchilas/gophermart/pkg/envflags"
	"time"
)

var (
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	LogLevel             string
	EnvType              string
	SecretKey            string
	WorkerPeriod         time.Duration
	WorkerPackLimit      int
	WorkerPoolSize       int
)

func ParseFlags() {
	flag.StringVar(&RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&DatabaseURI, "d", "", "database connection string")
	flag.StringVar(&AccrualSystemAddress, "r", "localhost:8081", "address of the accrual calculation service")
	flag.StringVar(&LogLevel, "l", "debug", "logging level")
	flag.StringVar(&EnvType, "e", "production", "type of environment (production or develop)")
	flag.StringVar(&SecretKey, "k", "supersecretkey", "the secret key for user tokens")
	flag.DurationVar(&WorkerPeriod, "w", 3*time.Second, "worker period of order enriching worker")
	flag.IntVar(&WorkerPackLimit, "p", 25, "pack limit of order enriching worker")
	flag.IntVar(&WorkerPoolSize, "z", 3, "pool size of order enriching worker")
	flag.Parse()

	envflags.TryUseEnvString(&RunAddress, "RUN_ADDRESS")
	envflags.TryUseEnvString(&DatabaseURI, "DATABASE_URI")
	envflags.TryUseEnvString(&AccrualSystemAddress, "ACCRUAL_SYSTEM_ADDRESS")
	envflags.TryUseEnvString(&LogLevel, "LOG_LEVEL")
	envflags.TryUseEnvString(&EnvType, "ENV_TYPE")
	envflags.TryUseEnvString(&SecretKey, "SECRET_KEY")
	envflags.TryUseEnvDuration(&WorkerPeriod, "WORKER_PERIOD")
	envflags.TryUseEnvInt(&WorkerPackLimit, "WORKER_PACK_LIMIT")
	envflags.TryUseEnvInt(&WorkerPoolSize, "WORKER_POOL_SIZE")
}
