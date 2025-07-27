package main

import (
	"log"

	"github.com/0x0FACED/sval"
)

type UserStorage struct {
	Users []User `sval:"user"`
}

type User struct {
	ID       string `sval:"id"`
	Name     string `sval:"name"`
	Age      int    `sval:"age"`
	Password string `sval:"password"`
	Email    string `sval:"email"`
}

func main() {
	// Example usage of UserStorage
	storage := UserStorage{
		Users: []User{
			{
				// correct
				ID:       "UID-TEST-123-END",
				Name:     "Alice",
				Age:      30,
				Password: "qBt3f!g6=-/f4h",
				Email:    "test@example.org",
			},
			{
				// incorrect id
				ID:       "UID-TEST-123-END-EXTRA",
				Name:     "Bob",
				Age:      25,
				Password: "n3g4s8f!g6=-/f4h",
				Email:    "test@example.org",
			},
			{
				// incorrect name
				ID:       "UID-TEST-123-END",
				Name:     "Im BOB 123!",
				Age:      25,
				Password: "n3g4s8f!g6=-/f4h",
				Email:    "test@example.org",
			},
			{
				// incorrect age
				ID:       "UID-TEST-123-END",
				Name:     "Im BOB 123!",
				Age:      8,
				Password: "n3g4s8f!g6=-/f4h",
				Email:    "test@example.org",
			},
			{
				// incorrect password
				ID:       "UID-TEST-123-END",
				Name:     "Im BOB 123!",
				Age:      25,
				Password: "12345",
				Email:    "test@example.org",
			},
			{
				// incorrect email
				ID:       "UID-TEST-123-END",
				Name:     "Im BOB 123!",
				Age:      25,
				Password: "n3g4s8f!g6=-/f4h",
				Email:    "test@@@example.com",
			},
			{
				// fully incorrect
				ID:       "Incorrect-ID",
				Name:     "Incorrect Name 123!",
				Age:      12,
				Password: "password123",
				Email:    "test@example@.com",
			},
		},
	}

	configLoader := &sval.FileConfigLoader{
		Path: "./examples/basic/sval.yaml",
	}

	validator, err := sval.NewWithConfig(configLoader)
	if err != nil {
		log.Fatalln("Failed to create validator:", err)
	}

	if err := validator.Validate(&storage); err != nil {
		log.Println(err.Error())
	} else {
		log.Println("All data is valid!")
	}
}
