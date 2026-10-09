// Package env reads settings from environment variables.
package env

import "os"

// Get returns the environment variable key, or fallback when it's unset.
func Get(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
