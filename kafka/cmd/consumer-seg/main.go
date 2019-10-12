package main

import (
	"context"
	"flag"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func getKafkaReader(kafkaURL, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaURL},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func main() {
	flag.Parse()

	kafkaURL := "127.0.0.1:9092"
	topic := "my_topic"
	groupID := "my-consumer-group"
	reader := getKafkaReader(kafkaURL, topic, groupID)
	defer reader.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-signals:
			return
		default:
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			msg, err := reader.ReadMessage(ctx)
			if err == context.DeadlineExceeded {
				continue
			} else if err != nil {
				log.Println(err)
				return
			}

			log.Printf("%s/%d/%d\t%s\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
		}
	}
}
