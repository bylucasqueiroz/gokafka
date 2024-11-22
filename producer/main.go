package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func main() {
	p, err := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers":  "localhost:29092",
			"enable.idempotence": true,
			"acks":               "all",
			"request.timeout.ms": 30000,
			"message.timeout.ms": 300000,
		},
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				// Check for error in delivery
				if ev.TopicPartition.Error != nil {
					// Log the error type for debugging
					log.Printf("Error type: %v\n", reflect.TypeOf(ev.TopicPartition.Error))

					// Check if the error is a kafka.Error type
					if kafkaErr, ok := ev.TopicPartition.Error.(kafka.Error); ok {
						// Check for specific error codes
						if kafkaErr.Code() == kafka.ErrTimedOut {
							log.Println("Timeout error occurred during delivery.")
						} else {
							log.Printf("Delivery failed: %v\n", kafkaErr)
						}
					} else {
						// Handle non-Kafka errors
						log.Printf("Non-Kafka error: %v\n", ev.TopicPartition.Error)
					}
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	topic := "test-topic"
	for _, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
	}

	p.Flush(3000)
}
