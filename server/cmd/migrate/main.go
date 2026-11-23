package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"accessible-content/server/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("migrate: переменная DATABASE_URL не задана")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("migrate: подключение: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("migrate: ping: %v", err)
	}

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("migrate: диалект: %v", err)
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	if err := goose.RunContext(ctx, command, db, "."); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}
