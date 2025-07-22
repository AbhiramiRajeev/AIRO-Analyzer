package config

import (
	"github.com/spf13/viper"
	"k8s.io/klog"
)

type Config struct {
	KafkaBrokers       []string       `mapstructure:"KAFKA_BROKERS"`
	KafkaTopic         string         `mapstructure:"KAFKA_TOPIC"`
	IncidentTopic      string         `mapstructure:"INCIDENT_TOPIC"`
	KafkaConsumerGroup string         `mapstructure:"KAFKA_CONSUMER_GROUP"`
	WindowSize         int            `mapstructure:"WINDOW_SIZE"`
	FailureThreshold   int            `mapstructure:"FAILURE_THRESHOLD"`
	PostgresConfig     PostgresConfig `mapstructure:",squash"` //squash allows you to embed the PostgresConfig struct directly into Config, so its fields are at the top level of Config
	RedisConfig        RedisConfig    `mapstructure:",squash"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"POSTGRES_HOST"`
	Port     int    `mapstructure:"POSTGRES_PORT"`
	Username string `mapstructure:"POSTGRES_USERNAME"`
	Password string `mapstructure:"POSTGRES_PASSWORD"`
	Database string `mapstructure:"POSTGRES_DATABASE"`
	SSLMode  string `mapstructure:"POSTGRES_SSL_MODE"`
}

type RedisConfig struct {
	Address  string `mapstructure:"REDIS_ADDRESS"`
	Password string `mapstructure:"REDIS_PASSWORD"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./.config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		klog.Info("Unable to read config file, using environment variables instead")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil { //Take all the configuration values it has loaded (from env, config file, and defaults), and populate your Go struct (cfg) with them.
		klog.Errorf("Unable to unmarshal config: %v", err)
		return nil, err
	}
	return &config, nil

}
