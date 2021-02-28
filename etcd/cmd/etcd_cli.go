package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"os"
	"strings"
	"time"
)

var callbacks = map[string]clientv3.WatchChan{}

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:23790", "localhost:23791", "localhost:23792"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	cli.KV = namespace.NewKV(cli.KV, "test/")
	cli.Watcher = namespace.NewWatcher(cli.Watcher, "test/")
	cli.Lease = namespace.NewLease(cli.Lease, "test/")

	go func() {
	callbacksLoop:
		for _, watch := range callbacks {
			select {
			case w := <-watch:
				for _, e := range w.Events {
					fmt.Printf("%s: %s (%s)\n", e.Kv.Key, e.Kv.Value, e.Type)
				}
			default:
				continue callbacksLoop
			}
		}
		time.Sleep(time.Second)
		goto callbacksLoop
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command> ")
		line, _, err := reader.ReadLine()
		if err != nil {
			panic(err)
		}

		if len(line) < 4 {
			fmt.Println("exit")
			break
		}
		cmd := string(line[:3])
		attrs := strings.Split(string(line[4:]), " ")

		switch cmd {
		case "get":
			var v string
			v, err = get(cli, attrs)
			fmt.Println(v)
		case "put":
			err = put(cli, attrs)
		case "del":
			err = del(cli, attrs)
		case "wch":
			watch(cli, attrs)
		default:
			fmt.Println("unexpected command")
		}

		if err != nil {
			fmt.Println(err)
		}
	}
}

func get(cli *clientv3.Client, attrs []string) (string, error) {
	resp, err := cli.Get(context.Background(), attrs[0])
	if err != nil {
		return "", err
	}

	for _, v := range resp.Kvs {
		return string(v.Value), nil
	}

	return "", nil
}

func del(cli *clientv3.Client, attrs []string) error {
	_, err := cli.Delete(context.Background(), attrs[0])
	return err
}
func watch(cli *clientv3.Client, attrs []string) {
	addCallback(attrs[0], cli.Watch(context.Background(), attrs[0]))
}
func put(cli *clientv3.Client, attrs []string) error {
	if len(attrs) != 2 {
		return fmt.Errorf("incorrect count arguments: %d, want 2", len(attrs))
	}

	_, err := cli.Put(context.Background(), attrs[0], attrs[1])
	return err
}

func addCallback(key string, watch clientv3.WatchChan) {
	if _, ok := callbacks[key]; ok {
		return
	}
	callbacks[key] = watch
}
