package internal

import (
	"database/sql"
	"log"

	"github.com/AbhiramiRajeev/Ingestion-Service/config"
	_ "github.com/lib/pq"
)

func InitDB(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("DB not reachable: %v", err)
	}
	log.Println("Connected to Postgres")
	return db
}
