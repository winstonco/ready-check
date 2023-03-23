package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	godotenv.Load("../.env")
	checkConfig()
}

func checkConfig() {
	var requiredVars = []string{
		"BOT_TOKEN",
	}
	for _, varName := range requiredVars {
		if _, exists := os.LookupEnv(varName); !exists {
			panic(fmt.Errorf("environment variable (%s) is undefined", varName))
		}
	}
}
