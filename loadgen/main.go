package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/weaveworks/common/server"
	"go.uber.org/ratelimit"
)

var (
	querierAddress   string
	queriesPerSecond int
	queryDurationMs  int
	workers          int

	latency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "loadgen",
		Name:      "latency",
		Help:      "latency",
		Buckets:   prometheus.LinearBuckets(0, .25, 20),
	}, []string{"tenant"})
)

func init() {
	flag.StringVar(&querierAddress, "querier-address", "http://localhost:3100", "Address to connect to for query frontend.")
	flag.IntVar(&queriesPerSecond, "queries-per-second", 10, "queries per second")
	flag.IntVar(&queryDurationMs, "query-duration-ms", 100, "query duration")
	flag.IntVar(&workers, "workers", 10, "query duration")
}

func main() {
	flag.Parse()

	servCfg := server.Config{
		GRPCListenPort:                 9095,
		HTTPListenPort:                 3100,
		ServerGracefulShutdownTimeout:  30 * time.Second,
		HTTPServerWriteTimeout:         30 * time.Second,
		HTTPServerReadTimeout:          30 * time.Second,
		HTTPServerIdleTimeout:          30 * time.Second,
		GPRCServerMaxRecvMsgSize:       10 * 1024 * 1024,
		GRPCServerMaxSendMsgSize:       10 * 1024 * 1024,
		GPRCServerMaxConcurrentStreams: 100,
		RegisterInstrumentation:        true,
	}
	serv, err := server.New(servCfg)
	if err != nil {
		panic(err)
	}

	duration := strconv.Itoa(queryDurationMs)
	limiter := ratelimit.New(queriesPerSecond)

	closeChan := make(chan struct{}, 0)

	tenantName, err := os.Hostname()
	if err != nil {
		fmt.Printf("hostname err: %v\n", err)
		panic(err)
	}

	client := &http.Client{Timeout: time.Second * 5}

	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
		L:
			for {
				select {
				case _, open := <-closeChan:
					if !open {
						break L
					}
				default:
					limiter.Take()
					req, err := http.NewRequest("POST", querierAddress, strings.NewReader(duration))
					if err != nil {
						fmt.Printf("newReq : %v\n", err)
						continue
					}

					req.Header.Set("X-Scope-OrgID", tenantName)

					start := time.Now()
					resp, err := client.Do(req)
					if err != nil {
						fmt.Printf("wups : %v\n", err)
						continue
					}
					latency.WithLabelValues(tenantName).Observe(time.Since(start).Seconds())
					resp.Body.Close()
				}

			}

			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		err = serv.Run()
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	close(closeChan)

	wg.Wait()
}
