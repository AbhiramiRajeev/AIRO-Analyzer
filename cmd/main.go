package main

import (
	"context"
	"log"

	"github.com/AbhiramiRajeev/AIRO-Analyzer/config"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/analyzer"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/db"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/kafka"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/redis"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisClient, err := redis.NewRedisClient(cfg.RedisConfig.Address, cfg.RedisConfig.Password)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	pgClient, err := db.NewRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer pgClient.Close()
	err = pgClient.CreateTable()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	defer pgClient.Close()

	analyzerService := analyzer.NewAnalyzerService(cfg, redisClient, *pgClient)

	handler := kafka.NewKafkaConsumerHandler(analyzerService)

	consumer, err := kafka.NewKafkaConsumer(cfg, handler)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	ctx := context.Background()
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Kafka consumer failed: %v", err)
	}
}
