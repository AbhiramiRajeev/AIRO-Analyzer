package config

import (
	"github.com/spf13/viper"
	"k8s.io/klog"
)


type Config struct {
	KafkaBrokers []string `mapstructure:"KAFKA_BROKERS"`
	KafkaTopic   string   `mapstructure:"KAFKA_TOPIC"`
	IncidentTopic string   `mapstructure:"INCIDENT_TOPIC"`
	KafkaConsumerGroup string `mapstructure:"KAFKA_CONSUMER_GROUP"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./.config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err!=nil {
		klog.Info("Unable to read config file, using environment variables instead")
	}

	var config Config
	if err:= viper.Unmarshal(&config); err!=nil { //Take all the configuration values it has loaded (from env, config file, and defaults), and populate your Go struct (cfg) with them.
		klog.Errorf("Unable to unmarshal config: %v", err)
		return nil, err
	}
	return &config, nil


}