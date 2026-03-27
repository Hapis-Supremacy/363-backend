package initializer

import (
	"log"

	"github.com/joho/godotenv"
)

// THis should only be called on dev
// Dont forget to remove .env before pushing, or ensure its untracked
// Request .env to Team Member
func LoadEnvVar() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
