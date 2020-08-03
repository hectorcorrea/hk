package tasks

import (
	"log"
	"os"
	"strings"

	"hectorcorrea.com/hk/models"
)

// Adds a new guest user.
func AddUser(userPassword string) {
	log.SetOutput(os.Stdout) // so we can redirect it

	tokens := strings.Split(userPassword, "/")
	if len(tokens) != 2 {
		log.Fatal("String must be in the form user/password")
	}

	if err := models.InitDB(); err != nil {
		log.Fatal("Failed to initialize database: %s", err)
	}
	log.Printf("Database: %s", models.DbConnStringSafe())

	err := models.AddGuestUser(tokens[0], tokens[1])
	if err != nil {
		log.Fatal("Error: %s", err)
	} else {
		log.Printf("Guest user %s added", tokens[0])
	}
}
