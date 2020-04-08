package main

import (
	"sync"
	"time"

	"github.com/joe-elliott/sim-query-frontend/worker"

	"github.com/cortexproject/cortex/pkg/querier/frontend"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/server"
)

func main() {
	servCfg := server.Config{
		GRPCListenPort: 9095,
		HTTPListenPort: 3100,
	}
	serv, err := server.New(servCfg)
	if err != nil {
		panic(err)
	}

	// start a query-frontend
	frontendCfg := frontend.Config{
		MaxOutstandingPerTenant: 100,
		CompressResponses:       true,
	}
	f, err := frontend.New(frontendCfg, util.Logger, prometheus.DefaultRegisterer)
	if err != nil {
		panic(err)
	}

	frontend.RegisterFrontendServer(serv.GRPC, f)
	serv.HTTP.PathPrefix("/").Handler(f.Handler())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = serv.Run()
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	workerCfg := worker.WorkerConfig{
		Address:           "localhost:9095",
		Parallelism:       10,
		DNSLookupDuration: 10 * time.Second,
	}
	w, err := worker.NewWorker(workerCfg, util.Logger)
	if err != nil {
		panic(err)
	}

	wg.Wait()
	w.Stopping(nil)
}
