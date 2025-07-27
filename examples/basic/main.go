package main

import (
	"encoding/json"
	"fmt"
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
	storage := UserStorage{
		Users: []User{
			{
				ID:       "UID-TEST-123-END", // Valid
				Name:     "Alice",            // Valid
				Age:      30,                 // Valid
				Password: "qBt3f!g6=-/f4h",   // Valid
				Email:    "test@example.org", // Valid
			},
			{
				ID:       "Invalid-ID",      // Invalid format
				Name:     "Bob123",          // Contains digits
				Age:      15,                // Too young
				Password: "weak",            // Too weak
				Email:    "bad@example.com", // Wrong domain
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
		// Pretty print the error
		data, _ := json.MarshalIndent(err, "", "  ")
		fmt.Println(string(data))
	}
}
