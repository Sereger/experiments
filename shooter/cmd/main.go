package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	rps = flag.Int("rps", 10, "rps")
	w   = flag.Int("w", 64, "workers")

	success int64
	fails   int64
)

type (
	fakeSrv struct {
		rnd *rand.Rand
	}

	requestor struct {
		url string
	}
)

func newTestServer() *httptest.Server {
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	return httptest.NewServer(&fakeSrv{rnd: rnd})
}

func (fs *fakeSrv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if fs.rnd.Intn(100) < 50 {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(500)
	}
}

func main() {
	flag.Parse()

	ch := make(chan *http.Request, *w)

	wg := sync.WaitGroup{}
	for i := 0; i < *w; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range ch {
				resp, err := http.DefaultClient.Do(r)
				if err != nil {
					atomic.AddInt64(&fails, 1)
					continue
				}

				if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
					atomic.AddInt64(&success, 1)
				} else {
					atomic.AddInt64(&fails, 1)
				}
				resp.Body.Close()
			}
		}()
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cnlFn := context.WithCancel(context.Background())
	go func() {
		<-sigCh
		cnlFn()
	}()

	srv := newTestServer()
	defer srv.Close()

	go printStat(ctx)

	rq := &requestor{url: srv.URL}
	startShoot(ctx, rq, *rps, ch)

	close(ch)

	wg.Wait()
}

func startShoot(ctx context.Context, rq *requestor, rps int, ch chan<- *http.Request) {
	d := time.Second / time.Duration(rps)
	tik := time.NewTicker(d)
	defer tik.Stop()

	for {
		select {
		case <-tik.C:
			ch <- rq.newRequest()
		case <-ctx.Done():
			return
		}
	}
}

func printStat(ctx context.Context) {
	tik := time.NewTicker(time.Second)
	defer tik.Stop()

	for {
		select {
		case <-tik.C:
			fmt.Printf("\rok %d / err %d", atomic.LoadInt64(&success), atomic.LoadInt64(&fails))
		case <-ctx.Done():
			fmt.Println()
			return
		}
	}
}

func (rq *requestor) newRequest() *http.Request {
	r, err := http.NewRequest(http.MethodGet, rq.url, nil)
	if err != nil {
		log.Panic(err)
	}

	return r
}
