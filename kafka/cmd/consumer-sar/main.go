package main

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	kafkaAddr = "127.0.0.1:9092"
	topik     = "my_topic"
)

func main() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	consumer := Consumer{
		ready: make(chan bool),
	}
	ctx, cnlFn := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup([]string{kafkaAddr}, "sag-group", config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}
	go func() {
		<-signals
		cnlFn()
	}()

	for {
		if err := client.Consume(ctx, []string{topik}, &consumer); err != nil {
			log.Printf("Error from consumer: %v", err)
		}
		// check if context was cancelled, signaling that the consumer should stop
		if ctx.Err() != nil {
			log.Printf("Error from consumer: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		default:
			continue
		}
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(ses sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for msg := range claim.Messages() {
		log.Printf("%s/%d/%d\t%s\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
		session.MarkMessage(msg, "")
	}

	return nil
}
