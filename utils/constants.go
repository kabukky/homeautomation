package utils

import "os"

var (
	HostAndPort        = envVarOrDefault("HOST_AND_PORT", ":2356")
	DirectoryDashboard = envVarOrDefault("DIRECTORY_DASHBOARD", "etc/dashboard")
)

func envVarOrDefault(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
