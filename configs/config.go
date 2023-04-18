package configs

import "github.com/joho/godotenv"

// loads the .env file
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	return nil
}
