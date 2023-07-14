package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"time"

	"google.golang.org/protobuf/proto"
	"telematics-generator/protobuf"
)

type Producer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

func (kp *Producer) ProduceMessage(telematicsData *protobuf.TelematicsDataProto) error {
	message, err := proto.Marshal(telematicsData)
	if err != nil {
		return err
	}

	err = kp.writer.WriteMessages(context.Background(), kafka.Message{
		Value: message,
	})
	if err != nil {
		return err
	}

	log.Printf("produced message: %s", telematicsData.String())
	return nil
}

func (kp *Producer) Close() error {
	return kp.writer.Close()
}
