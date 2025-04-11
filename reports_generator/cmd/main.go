package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/KostySCH/Reports_go/reports_generator/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connStr := "postgres://postgres:kosty8021@localhost:5432/MedicalClinic?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Not connect: %v\n", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepo(pool)

	ids, err := repo.GetUserIDs(context.Background())
	if err != nil {
		log.Fatalf("Error get user id: %v\n", err)
	}

	if len(ids) == 0 {
		fmt.Println("404")
		os.Exit(0)
	}

	fmt.Println("User id:")
	for _, id := range ids {
		fmt.Printf("%d ", id)
	}
	fmt.Println()
}
