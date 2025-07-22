package kafka

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/AbhiramiRajeev/AIRO-Analyzer/config"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/analyzer"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/models"
	"github.com/IBM/sarama"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog"
)

// KafkaConsumer wraps Sarama consumer group
type KafkaConsumer struct {
	Group   sarama.ConsumerGroup
	Topic   string
	Handler sarama.ConsumerGroupHandler
}

// NewKafkaConsumer creates a new consumer group client
func NewKafkaConsumer(cfg *config.Config, handler sarama.ConsumerGroupHandler) (*KafkaConsumer, error) {
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

func NewKafkaConsumerHandler(analyzer *analyzer.AnalyzerService) sarama.ConsumerGroupHandler {
	return &KafkaConsumerHandler{analyzer: analyzer}
}

// Start begins consuming messages and handles graceful shutdown
func (kc *KafkaConsumer) Start(ctx context.Context) error {

	go func() {
		for {
			if err := kc.Group.Consume(ctx, []string{kc.Topic}, kc.Handler); err != nil {
				klog.Info("Error while consuming %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}

	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT) //SIGINT - press Ctrl + C in the terminal running your Go program SIGTERM - signal used by operating systems to request that a program terminate gracefull (Kubernetes Docker)
	<-sig                                               // wait for the signal
	klog.Info("Shuttimg down the consumer")
	return kc.Group.Close() // After receiving the signal (<-sig), your program is likely exiting or cleaning up and exiting. So closing sig is unnecessary, and doing so manually is even discouraged unless:

}

//////////////////////////////////////////////////////////////
// KafkaConsumerHandler implements sarama.ConsumerGroupHandler
//////////////////////////////////////////////////////////////

type KafkaConsumerHandler struct {
	analyzer *analyzer.AnalyzerService
}

func (kh *KafkaConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (kh *KafkaConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (kh *KafkaConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// This method is called when a new claim is made by the consumer group.
	// It will be called for each partition that the consumer group is assigned to.
	// You can implement your logic here to process messages from the claim.
	// For example, you can use a loop to read messages from the claim and process them.
	// You can also use the analyzer service to analyze the messages.
	// This method should return an error if there is an issue processing the claim.
	for msg := range claim.Messages() {

		var event models.Event
		err := json.Unmarshal(msg.Value, &event)
		if err != nil {
			klog.Error("Error unmarshalling message: ", err)
			session.MarkMessage(msg, "error")
			continue
		}

		err = kh.analyzer.Analyze(event)
		if err != nil {
			klog.Error("Error analyzing event: ", err)
			session.MarkMessage(msg, "error")
			continue
		}
		session.MarkMessage(msg, "") // Mark the message as processed
	}
	return nil

}
