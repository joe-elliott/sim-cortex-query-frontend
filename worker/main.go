package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/grpcclient"
)

func main() {
	// worker
	workerCfg := WorkerConfig{
		Address:           "localhost:9095",
		Parallelism:       10,
		DNSLookupDuration: 10 * time.Second,
		GRPCClientConfig: grpcclient.Config{
			MaxRecvMsgSize: 10 * 1024 * 1024,
			MaxSendMsgSize: 10 * 1024 * 1024,
		},
	}
	w, err := NewWorker(workerCfg, util.Logger)
	if err != nil {
		panic(err)
	}

	// start worker
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err = w.WatchDNSLoop(ctx)
		if err != nil {
			panic(err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	cancel()
	w.Stopping(nil)
}
