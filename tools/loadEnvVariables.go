package tools

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func LoadEnvVariables()  {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
func GetEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0
	}

	return valueInt
}

