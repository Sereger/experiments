package main

import (
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := cluster.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Consumer.Return.Errors = false
	config.Consumer.IsolationLevel = sarama.ReadCommitted
	//config.Consumer.Fetch.Default = 1
	//config.Consumer.Fetch.Max = 1
	//config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	//config.Group.PartitionStrategy = cluster.StrategyRoundRobin
	//config.Group.Mode = cluster.ConsumerModeMultiplex

	// init consumer
	brokers := []string{"127.0.0.1:9092"}
	topics := []string{"my_topic"}
	consumer, err := cluster.NewConsumer(brokers, "my-consumer-group", topics, config)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// consume partitions
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				consumer.MarkOffset(msg, "") // mark message as processed
				log.Printf("%s/%d/%d\t%s\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
			}
		case <-signals:
			return
		}
	}
}
