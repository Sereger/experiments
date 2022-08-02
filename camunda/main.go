package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	camunda_client_go "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
	"github.com/google/uuid"
)

type loggedTransport struct{}

func (lt *loggedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	wrappedReq := io.TeeReader(r.Body, os.Stdout)
	r.Body = io.NopCloser(wrappedReq)
	return http.DefaultTransport.RoundTrip(r)
}

var verbose = flag.Bool("v", false, "")

func main() {
	flag.Parse()

	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: "http://localhost:8080/engine-rest",
		ApiUser:     "demo",
		ApiPassword: "demo",
		Timeout:     time.Second * 10,
	})
	if *verbose {
		client.SetCustomTransport(new(loggedTransport))
	}

	logger := func(err error) {
		fmt.Println(err.Error())
	}
	proc := processor.NewProcessor(client, &processor.Options{
		WorkerId:                  "process-worker",
		LockDuration:              time.Minute * 2,
		MaxTasks:                  10,
		MaxParallelTaskPerHandler: 100,
		LongPollingTimeout:        5 * time.Second,
	}, logger)

	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	proc.AddHandler(
		[]*camunda_client_go.QueryFetchAndLockTopic{
			{TopicName: "process-uber"},
			{TopicName: "process-glovo"},
		},
		handlerWrapper(processPartners, rnd),
	)
	proc.AddHandler(
		[]*camunda_client_go.QueryFetchAndLockTopic{
			{TopicName: "register-order"},
		},
		handlerWrapper(register),
	)
	proc.AddHandler(
		[]*camunda_client_go.QueryFetchAndLockTopic{
			{TopicName: "make-receipt"},
			{TopicName: "process-cancellation"},
		},
		handlerWrapper(nil),
	)

	pickedCh := make(chan pickerJob, 32)
	proc.AddHandler(
		[]*camunda_client_go.QueryFetchAndLockTopic{
			{TopicName: "picking"},
		},
		handlerWrapper(picking, pickedCh),
	)
	refundCh := make(chan refundJob, 32)
	proc.AddHandler(
		[]*camunda_client_go.QueryFetchAndLockTopic{
			{TopicName: "call-to-user"},
		},
		handlerWrapper(callToUser, refundCh),
	)

	reader := bufio.NewReader(os.Stdin)
	for i := 0; i < 4; i++ {
		go func() {
			for picked := range pickedCh {
				time.Sleep(2 * time.Second)
				fmt.Printf("how many items picked for order [%s] (total %d)\n", picked.orderID, picked.itemsCount)
				text, _, _ := reader.ReadLine()
				count, _ := strconv.Atoi(string(text))
				debug("inputed: %d (%s)\n", count, text)
				status := "picked"
				if picked.itemsCount != int64(count) {
					status = "partly_picked"
				}
				params := map[string]camunda_client_go.Variable{
					"picked_items": {
						Value: count,
						Type:  "integer",
					},
					"order_id": {
						Value: picked.orderID,
						Type:  "string",
					},
					"status": {
						Value: status,
						Type:  "string",
					},
				}
				err := client.Message.SendMessage(&camunda_client_go.ReqMessage{
					MessageName:      "order-picked",
					BusinessKey:      picked.businessKey,
					ProcessVariables: &params,
				})
				if err != nil {
					fmt.Printf("send message err: %s (%s)\n", err, picked.orderID)
				}
			}
		}()
	}

	for i := 0; i < 4; i++ {
		go func() {
			for refund := range refundCh {
				time.Sleep(time.Second)
				fmt.Printf(
					"picked %d of %d, original total: %d cents, what the new total?\n",
					refund.itemsPicked, refund.itemsCount, refund.total,
				)
				text, _, _ := reader.ReadLine()
				total, _ := strconv.Atoi(string(text))
				debug("(debug) inputted: %d (%s)\n", total, text)
				next := "order-refund"
				status := "picked"
				if total == 0 {
					next = "order-canceled"
					status = "canceled"
				}

				params := map[string]camunda_client_go.Variable{
					"total": {
						Value: total,
						Type:  "integer",
					},
					"order_id": {
						Value: refund.orderID,
						Type:  "string",
					},
					"status": {
						Value: status,
						Type:  "string",
					},
				}

				err := client.Message.SendMessage(&camunda_client_go.ReqMessage{
					MessageName:      next,
					BusinessKey:      refund.businessKey,
					ProcessVariables: &params,
				})
				if err != nil {
					fmt.Printf("send message err: %s (%s)\n", err, refund.orderID)
				}
			}
		}()
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	proc.Shutdown()

	history, err := client.History.GetProcessInstanceList(nil)
	if err != nil {
		fmt.Printf("history err: %s\n", err)
		return
	}
	for _, h := range history {
		fmt.Printf("%s | %s | %s | %s\n", h.BusinessKey, h.StartTime, h.EndTime, h.State)
	}
}

type (
	handler = func(*processor.Context, ...interface{}) error
)

func handlerWrapper(fn handler, args ...interface{}) func(*processor.Context) error {
	return func(ctx *processor.Context) error {
		fmt.Printf("Running task %s. WorkerId: %s. TopicName: %s\n", ctx.Task.Id, ctx.Task.WorkerId, ctx.Task.TopicName)
		debug("%+v\n", ctx.Task.Variables)
		defer debug("Task %s completed\n", ctx.Task.Id)

		if fn == nil {
			return complete(ctx, ctx.Task.Variables)
		}
		err := fn(ctx, args...)
		if err != nil {
			fmt.Printf("Task %s err: %s\n", ctx.Task.Id, err.Error())
		}

		return err
	}
}

func processPartners(ctx *processor.Context, args ...interface{}) error {
	if v, ok := ctx.Task.Variables["error"]; ok && v.Value.(bool) {
		err := "partner is not available"
		details := "some details"
		return ctx.HandleFailure(processor.QueryHandleFailure{
			ErrorMessage: &err,
			ErrorDetails: &details,
		})
	}

	source := ctx.Task.Variables["source"].Value.(string)
	var fee int64
	switch source {
	case "uber":
		fee = 299
	case "glovo":
		fee = 999
	}
	total := int64(ctx.Task.Variables["total"].Value.(float64))
	ctx.Task.Variables["fee"] = camunda_client_go.Variable{
		Value: fee,
		Type:  "integer",
	}
	ctx.Task.Variables["total"] = camunda_client_go.Variable{
		Value: total + fee,
		Type:  "integer",
	}

	return complete(ctx, ctx.Task.Variables)
}
func register(ctx *processor.Context, _ ...interface{}) error {
	data := ctx.Task.Variables["additionalJsonInfoObject"].Value.(string)
	type (
		orderItem struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Qty   int64  `json:"qty"`
			Price int64  `json:"price"`
		}
		order struct {
			OrderID string      `json:"order_id"`
			Source  string      `json:"source"`
			Items   []orderItem `json:"items"`
			Total   int64       `json:"total"`
		}
	)
	o := &order{}
	err := json.Unmarshal([]byte(data), o)
	if err != nil {
		return err
	}

	ctx.Task.Variables = map[string]camunda_client_go.Variable{
		"internal_id": {
			Value:     uuid.New().String(),
			Type:      "string",
			ValueInfo: camunda_client_go.ValueInfo{},
		},
		"order_id": {
			Value: o.OrderID,
			Type:  "string",
		},
		"source": {
			Value: o.Source,
			Type:  "string",
		},
		"items_count": {
			Value: len(o.Items),
			Type:  "integer",
		},
		"total": {
			Value: o.Total,
			Type:  "integer",
		},
		"status": {
			Value: "new",
			Type:  "string",
		},
	}
	if strings.HasPrefix(ctx.Task.BusinessKey, "err") {
		ctx.Task.Variables["error"] = camunda_client_go.Variable{
			Value: true,
			Type:  "boolean",
		}
	}
	return complete(ctx, ctx.Task.Variables)
}

func complete(ctx *processor.Context, v map[string]camunda_client_go.Variable) error {
	err := ctx.Complete(processor.QueryComplete{
		Variables: &v,
	})
	if err != nil {
		fmt.Printf("Error set complete task %s: %s\n", ctx.Task.Id, err)
		errTxt := err.Error()
		return ctx.HandleFailure(processor.QueryHandleFailure{
			ErrorMessage: &errTxt,
		})
	}
	return nil
}

type pickerJob struct {
	orderID     string
	itemsCount  int64
	businessKey string
}

func picking(ctx *processor.Context, args ...interface{}) error {
	ch := args[0].(chan pickerJob)
	ch <- pickerJob{
		orderID:     ctx.Task.Variables["order_id"].Value.(string),
		itemsCount:  int64(ctx.Task.Variables["items_count"].Value.(float64)),
		businessKey: ctx.Task.BusinessKey,
	}
	return complete(ctx, ctx.Task.Variables)
}

type refundJob struct {
	orderID     string
	businessKey string
	itemsCount  int64
	itemsPicked int64
	total       int64
}

func callToUser(ctx *processor.Context, args ...interface{}) error {
	ch := args[0].(chan refundJob)
	ch <- refundJob{
		orderID:     ctx.Task.Variables["order_id"].Value.(string),
		businessKey: ctx.Task.BusinessKey,
		itemsCount:  int64(ctx.Task.Variables["items_count"].Value.(float64)),
		itemsPicked: int64(ctx.Task.Variables["picked_items"].Value.(float64)),
		total:       int64(ctx.Task.Variables["total"].Value.(float64)),
	}

	return complete(ctx, ctx.Task.Variables)
}

func debug(format string, a ...any) {
	if *verbose {
		fmt.Printf(format, a...)
	}
}
