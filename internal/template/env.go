package template

import (
	"fmt"
	"os"
)

func env(name string) (string, error) {
	if value := os.Getenv(name); value == "" {
		return "", fmt.Errorf("env variable %q was empty", name)
	} else {
		return value, nil
	}
}

func envDefault(name, fallback string) (string, error) {
	if value := os.Getenv(name); value == "" {
		return fallback, nil
	} else {
		return value, nil
	}
}
