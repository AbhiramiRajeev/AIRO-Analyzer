package kafka

import (
	"github.com/AbhiramiRajeev/Ingestion-Service/config"
	"github.com/IBM/sarama"
	"k8s.io/klog"
)

type KafkaProducer struct {
	Producer sarama.SyncProducer
	topic   string
}

func NewKafkaProducer(cfg *config.Config) (*KafkaProducer, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 5
	producer , err := sarama.NewSyncProducer(cfg.KafkaBrokers, kafkaConfig)
	if err != nil {
		klog.Info("Error creating Kafka producer: ", err)
		return nil, err
	}
	return &KafkaProducer{
		Producer: producer,
		topic:   cfg.KafkaTopic,
	} , nil

}

func (k *KafkaProducer) Publish(message []byte) error{
	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.ByteEncoder(message),	
	}
	partition , offset , err := k.Producer.SendMessage(msg)
	if err != nil {
		klog.Error("Error sending message to Kafka: ", err)
		return err
	}
	klog.Infof("Message sent to partition %d at offset %d", partition, offset)
	return nil
}