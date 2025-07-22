package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/AbhiramiRajeev/AIRO-Analyzer/config"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/models"
	_ "github.com/lib/pq"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(cfg *config.Config) (*Repository, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.PostgresConfig.Host, cfg.PostgresConfig.Port, cfg.PostgresConfig.Username, cfg.PostgresConfig.Password, cfg.PostgresConfig.Database, cfg.PostgresConfig.SSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
		return nil, err
	}
	fmt.Println("Connected to the database successfully")
	return &Repository{
		DB: db,
	}, nil

}

func (r *Repository) Close() error {
	if err := r.DB.Close(); err != nil {
		return fmt.Errorf("error closing database connection: %v", err)
	}
	return nil
}
func (r *Repository) CreateTable() error {
	query :=
		`CREATE TABLE IF NOT EXISTS incidents(
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) NOT NULL,
        ip VARCHAR(255) NOT NULL,
        timestamp TIMESTAMP NOT NULL,
        incident_type VARCHAR(50) NOT NULL,
        description TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`

	if _, err := r.DB.Exec(query); err != nil {
		return fmt.Errorf("error creating incidents table: %v", err)
	}
	return nil
}
func (r *Repository) AddIncident(incident models.Incident) error {
	query := `INSERT INTO incidents (username, ip, timestamp, incident_type, description) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(query, incident.UserID, incident.IpAddress, incident.Timestamp, incident.EventType, incident.Details)
	if err != nil {
		return fmt.Errorf("error inserting incident: %v", err)
	}
	return nil

}
