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

	"go.uber.org/ratelimit"
)

var (
	querierAddress   string
	queriesPerSecond int
	queryDurationMs  int
	workers          int
)

func init() {
	flag.StringVar(&querierAddress, "querier-address", "http://localhost:3100", "Address to connect to for query frontend.")
	flag.IntVar(&queriesPerSecond, "queries-per-second", 10, "queries per second")
	flag.IntVar(&queryDurationMs, "query-duration-ms", 100, "query duration")
	flag.IntVar(&workers, "workers", 10, "query duration")
}

func main() {
	flag.Parse()

	duration := strconv.Itoa(queryDurationMs)
	limiter := ratelimit.New(queriesPerSecond)

	closeChan := make(chan struct{}, 0)

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
					_, err := http.Post(querierAddress, "text", strings.NewReader(duration))
					if err != nil {
						fmt.Printf("wups : %v\n", err)
					}
				}

			}

			wg.Done()
		}()
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	close(closeChan)

	wg.Wait()
}
