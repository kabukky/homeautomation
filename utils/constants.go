package utils

import (
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	APIBasePath = "/api/v1/"

	HostAndPort        = envVarOrDefault("HOST_AND_PORT", ":2356")
	DirectoryDashboard = envVarOrDefault("DIRECTORY_DASHBOARD", "etc/dashboard")

	WeatherLatitude      = mustFloatEnvVar("WEATHER_LATITUDE", "53.900149")
	WeatherLongitude     = mustFloatEnvVar("WEATHER_LONGITUDE", "10.739025")
	OpenWeatherMapAPIKey = mustEnvVar("OPENWEATHERMAP_API_KEY")

	GoogleCalendarIDs = strings.Split(envVarOrDefault("GOOGLE_CALENDAR_IDS", ""), ",")
)

func mustEnvVar(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("Required env var %s is not set", name)
	}
	return value
}

func mustFloatEnvVar(name, defaultValue string) float32 {
	result, err := strconv.ParseFloat(envVarOrDefault(name, defaultValue), 32)
	if err != nil {
		log.Fatalf("Could not parse env var %s as float", name)
	}
	return float32(result)
}

func envVarOrDefault(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
