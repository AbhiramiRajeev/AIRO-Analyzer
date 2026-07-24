package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func viperReset() {
	viper.Reset()
}

func TestLoadConfigFromFile(t *testing.T) {
	content := `
KAFKA_BROKERS:
  - localhost:9092
KAFKA_TOPIC: test_events
INCIDENT_TOPIC: test_incidents
KAFKA_CONSUMER_GROUP: test-group
WINDOW_SIZE: 300
FAILURE_THRESHOLD: 3
POSTGRES_HOST: localhost
POSTGRES_PORT: 5432
POSTGRES_USERNAME: testuser
POSTGRES_PASSWORD: testpass
POSTGRES_DATABASE: testdb
POSTGRES_SSL_MODE: disable
REDIS_ADDRESS: localhost:6379
REDIS_PASSWORD: ""
`
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Reset viper state
	viperReset()

	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if len(cfg.KafkaBrokers) != 1 || cfg.KafkaBrokers[0] != "localhost:9092" {
		t.Errorf("KafkaBrokers: got %v, want [localhost:9092]", cfg.KafkaBrokers)
	}
	if cfg.KafkaTopic != "test_events" {
		t.Errorf("KafkaTopic: got %s, want test_events", cfg.KafkaTopic)
	}
	if cfg.IncidentTopic != "test_incidents" {
		t.Errorf("IncidentTopic: got %s, want test_incidents", cfg.IncidentTopic)
	}
	if cfg.KafkaConsumerGroup != "test-group" {
		t.Errorf("KafkaConsumerGroup: got %s, want test-group", cfg.KafkaConsumerGroup)
	}
	if cfg.WindowSize != 300 {
		t.Errorf("WindowSize: got %d, want 300", cfg.WindowSize)
	}
	if cfg.FailureThreshold != 3 {
		t.Errorf("FailureThreshold: got %d, want 3", cfg.FailureThreshold)
	}
	if cfg.PostgresConfig.Host != "localhost" {
		t.Errorf("PostgresConfig.Host: got %s, want localhost", cfg.PostgresConfig.Host)
	}
	if cfg.PostgresConfig.Port != 5432 {
		t.Errorf("PostgresConfig.Port: got %d, want 5432", cfg.PostgresConfig.Port)
	}
	if cfg.PostgresConfig.Username != "testuser" {
		t.Errorf("PostgresConfig.Username: got %s, want testuser", cfg.PostgresConfig.Username)
	}
	if cfg.PostgresConfig.Password != "testpass" {
		t.Errorf("PostgresConfig.Password: got %s, want testpass", cfg.PostgresConfig.Password)
	}
	if cfg.PostgresConfig.Database != "testdb" {
		t.Errorf("PostgresConfig.Database: got %s, want testdb", cfg.PostgresConfig.Database)
	}
	if cfg.PostgresConfig.SSLMode != "disable" {
		t.Errorf("PostgresConfig.SSLMode: got %s, want disable", cfg.PostgresConfig.SSLMode)
	}
	if cfg.RedisConfig.Address != "localhost:6379" {
		t.Errorf("RedisConfig.Address: got %s, want localhost:6379", cfg.RedisConfig.Address)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("config should not be nil")
	}
}
