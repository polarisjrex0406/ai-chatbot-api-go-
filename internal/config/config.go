package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GetOpenAiApiKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

func GetNetziloApiKey() string {
	return os.Getenv("NETZILO_API_KEY")
}
