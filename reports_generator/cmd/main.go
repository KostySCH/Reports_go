package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/KostySCH/Reports_go/reports_generator/internal/repository"
)

func main() {
	connStr := "postgres://postgres:kosty8021@localhost:5432/MedicalClinic?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepo(pool)

	ids, err := repo.GetUserIDs(context.Background())
	if err != nil {
		log.Fatalf("Failed to get user IDs: %v\n", err)
	}

	if len(ids) == 0 {
		fmt.Println("No users found in database")
		os.Exit(0)
	}

	fmt.Println("User IDs from database:")
	for _, id := range ids {
		fmt.Printf("%d ", id)
	}
	fmt.Println()
}