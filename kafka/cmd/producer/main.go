package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	kafkaAddr = "127.0.0.1:9092"
	topik     = "my_topic"
)

func main() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0

	config.Producer.Return.Successes = true

	err := createTopik(kafkaAddr, config, topik, 3)
	if err != nil {
		log.Fatal(err)
	}

	producer, err := sarama.NewAsyncProducer([]string{kafkaAddr}, config)
	if err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var (
		wg                          sync.WaitGroup
		enqueued, successes, errors int
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		logChars := []string{".", ":", "|"}
		for msg := range producer.Successes() {
			successes++
			fmt.Print(logChars[msg.Partition])
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			log.Println(err)
			errors++
		}
	}()

ProducerLoop:
	for {
		message := &sarama.ProducerMessage{
			Topic: "my_topic",
			Value: sarama.StringEncoder("testing " + strconv.Itoa(enqueued)),
		}
		select {
		case producer.Input() <- message:
			enqueued++
			time.Sleep(2 * time.Second)
		case <-signals:
			producer.AsyncClose() // Trigger a shutdown of the producer.
			break ProducerLoop
		}
	}
	fmt.Println()

	wg.Wait()

	log.Printf("Successfully produced: %d; errors: %d\n", successes, errors)
}

func createTopik(addr string, config *sarama.Config, topic string, numPartitions int32) error {
	broker := sarama.NewBroker(addr)
	err := broker.Open(config)
	if err != nil {
		return err
	}
	// check if the connection was OK
	_, err = broker.Connected()
	if err != nil {
		return err
	}

	topicDetail := &sarama.TopicDetail{}
	topicDetail.NumPartitions = numPartitions
	topicDetail.ReplicationFactor = int16(1)
	topicDetail.ConfigEntries = make(map[string]*string)

	topicDetails := make(map[string]*sarama.TopicDetail)
	topicDetails[topic] = topicDetail

	request := sarama.CreateTopicsRequest{
		Timeout:      time.Second * 15,
		TopicDetails: topicDetails,
	}

	// Send request to Broker
	_, err = broker.CreateTopics(&request)

	return err
}
