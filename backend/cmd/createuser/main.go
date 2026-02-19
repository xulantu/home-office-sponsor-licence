package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"sponsor-tracker/internal/config"
	"sponsor-tracker/internal/database"
)

func main() {
	role := flag.Int("role", 10, "user role (10=admin, 50=viewer)")
	flag.Parse()

	cfg, err := config.Load("config.yaml", ".env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := database.Connect(cfg.Database.ConnectionString())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	var username, password string
	fmt.Print("Username: ")
	if _, err := fmt.Scanln(&username); err != nil {
		log.Fatalf("failed to read username: %v", err)
	}
	fmt.Print("Password: ")
	if _, err := fmt.Scanln(&password); err != nil {
		log.Fatalf("failed to read password: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	user := database.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         *role,
	}
	id, err := database.InsertUser(context.Background(), pool, user)
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	fmt.Printf("Created user %q with ID %d and role %d\n", username, id, *role)
}
