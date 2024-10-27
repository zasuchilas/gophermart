package envflags

import "os"

func TryUseEnvString(flagValue *string, envName string) {
	if env := os.Getenv(envName); env != "" {
		*flagValue = env
	}
}
