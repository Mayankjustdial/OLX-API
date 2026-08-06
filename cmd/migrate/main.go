package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codersgyan/olx-api/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up | down>")
	}

	cfg := config.MustLoad()

	m, err := migrate.New(
		"file://migrations",
		cfg.DatabaseUrl,
	)

	if err != nil {
		log.Fatalf("migration.new: %v", err)
	}

	switch os.Args[1] {
	case "up":
		log.Printf("up called")
		if err := m.Up(); err != nil {
			log.Fatal(err)
		}

	case "down":
		log.Printf("down called")
		if err := m.Down(); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("unknown command")

	}

	fmt.Println("running migration")
}
