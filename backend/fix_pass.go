package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	
	_, err = dbPool.Exec(context.Background(), "UPDATE users SET password_hash = $1 WHERE email = 'dinesh@test.com'", string(hash))
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Password updated successfully to a valid bcrypt hash!")
}
