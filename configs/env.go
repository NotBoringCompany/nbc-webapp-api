package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// loads the .env file
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	return nil
}

// specifically loads the mongodb URI from the .env file
func LoadMongoEnv() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file to retrieve MongoDB URI")
	}

	return os.Getenv("MONGODB_URI")
}
