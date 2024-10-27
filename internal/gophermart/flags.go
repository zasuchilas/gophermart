package gophermart

import (
	"flag"
	"github.com/zasuchilas/gophermart/pkg/envflags"
	"os"
)

var (
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
)

func parseFlags() {
	flag.StringVar(&RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&DatabaseURI, "d", "", "database connection string")
	flag.StringVar(&AccrualSystemAddress, "r", "localhost:8081", "address of the accrual calculation service")
	flag.Parse()

	envflags.TryUseEnvString(&RunAddress, "RUN_ADDRESS")
	envflags.TryUseEnvString(&DatabaseURI, "DATABASE_URI")
	envflags.TryUseEnvString(&AccrualSystemAddress, "ACCRUAL_SYSTEM_ADDRESS")
	if env := os.Getenv("RUN_ADDRESS"); env != "" {
		RunAddress = env
	}
	if env := os.Getenv("DATABASE_URI"); env != "" {
		DatabaseURI = env
	}
	if env := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); env != "" {
		AccrualSystemAddress = env
	}
}
