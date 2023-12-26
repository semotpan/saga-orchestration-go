package kafka

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.example/saga/payment/pkg/model"
	"log"
)

// Ingester defines a Kafka ingester.
type Ingester struct {
	consumer *kafka.Consumer
	topic    string
}

// NewIngester creates a new Kafka ingester.
func NewIngester(addr string, groupID string, topic string) (*Ingester, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": addr,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	return &Ingester{consumer, topic}, nil
}

// Ingest starts ingestion from Kafka and returns a channel containing rating events
func (i *Ingester) Ingest(ctx context.Context) (chan model.PaymentEvent, error) {
	if err := i.consumer.SubscribeTopics([]string{i.topic}, nil); err != nil {
		return nil, err
	}

	ch := make(chan model.PaymentEvent, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				i.consumer.Close()
				return
			default:
			}
			msg, err := i.consumer.ReadMessage(-1)
			if err != nil {
				log.Println("Consumer error: " + err.Error())
				continue
			}
			var payload model.Payment
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Println("Unmarshal error: " + err.Error())
				continue
			}
			var eventId string
			for _, v := range msg.Headers {
				if v.Key == "id" {
					eventId = string(v.Value)
					break
				}
			}
			ch <- model.PaymentEvent{
				EventID:   eventId,
				MsgID:     string(msg.Key),
				Timestamp: msg.Timestamp,
				Payload:   payload,
			}
		}
	}()
	return ch, nil
}
