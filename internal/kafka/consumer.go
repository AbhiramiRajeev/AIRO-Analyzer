package kafka

import (
	"context"
	"encoding/json"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/analyzer"
	"github.com/AbhiramiRajeev/Ingestion-Service/config"
	"github.com/IBM/sarama"
	"k8s.io/klog"
	"os"
	"os/signal"
	"syscall"
)

// KafkaConsumer wraps Sarama consumer group
type KafkaConsumer struct {
	Group   sarama.ConsumerGroup
	Topic   string
	Handler sarama.ConsumerGroupHandler
}

// NewKafkaConsumer creates a new consumer group client
func NewKafkaConsumer(cfg config.Config, handler sarama.ConsumerGroupHandler) (*KafkaConsumer, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true // This allows your app to monitor and log Kafka consumption errors like connection issues, topic not found, etc.
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange() 
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaConsumerGroup, kafkaConfig)
	if err != nil {
		klog.Error("Failed to create Kafka consumer group: ", err)
		return nil, err
	}

	return &KafkaConsumer{
		Group:   group,
		Topic:   cfg.KafkaTopic,
		Handler: handler,
	}, nil
}

// Start begins consuming messages and handles graceful shutdown
func (kc *KafkaConsumer)Start(ctx context.Context)error	{

	go func() {
		for {
			err := kc.Group.Consume(ctx,[]string{kc.Topic},kc.Handler); err!=nil {
				klog.Info("Error while consuming %v",err)
			}

			if ctx.Err()!=nil {
				return
			}
		}

	}()
	
	return nil



}

//////////////////////////////////////////////////////////////
// KafkaConsumerHandler implements sarama.ConsumerGroupHandler
//////////////////////////////////////////////////////////////

type KafkaConsumerHandler struct {
	Analyzer *analyzer.AnalyzerService
}

// Setup is run before starting to consume, can be used for init
func (h *KafkaConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run after consumer stops
func (h *KafkaConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from the Kafka topic
func (h *KafkaConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		klog.Infof("Message claimed: topic = %s, partition = %d, offset = %d", msg.Topic, msg.Partition, msg.Offset)

		var loginEvent analyzer.LoginEvent
		err := json.Unmarshal(msg.Value, &loginEvent)
		if err != nil {
			klog.Errorf("Failed to unmarshal message: %v", err)
			continue
		}

		// Call the analyzer logic
		err = h.Analyzer.ProcessLoginEvent(&loginEvent)
		if err != nil {
			klog.Errorf("Analyzer failed to process event: %v", err)
		}

		// Mark message as processed
		session.MarkMessage(msg, "")
	}
	return nil
}
