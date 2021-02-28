package main

import (
	"context"
	"fmt"
	"github.com/Sereger/experiments/etcd"
	"time"
)

type cfg struct {
	Host string
	Port int64
	Keys []string
}

func main() {
	w, err := etcd.New([]string{"localhost:23790", "localhost:23791", "localhost:23792"}, "test", 5*time.Second)
	if err != nil {
		panic(err)
	}
	cfg := &cfg{
		Host: "localhost",
		Port: 54321,
	}
	err = w.SyncConfig(context.Background(), cfg, func(key string, newValue interface{}) {
		fmt.Printf("updated value %s: %s\n", key, newValue)
		fmt.Printf("%+v\n", cfg)
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", cfg)
	for {
		time.Sleep(time.Second)
	}
}
