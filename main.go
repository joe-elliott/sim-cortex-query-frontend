package main

import (
	"context"
	"sync"
	"time"

	"net/http"

	"github.com/joe-elliott/sim-query-frontend/worker"

	"github.com/cortexproject/cortex/pkg/querier/frontend"
	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/grpcclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/middleware"
	"github.com/weaveworks/common/server"
	"github.com/weaveworks/common/user"
)

func main() {
	// server
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

	// frontend
	frontendCfg := frontend.Config{
		MaxOutstandingPerTenant: 100,
		CompressResponses:       true,
	}
	f, err := frontend.New(frontendCfg, util.Logger, prometheus.DefaultRegisterer)
	if err != nil {
		panic(err)
	}

	authMiddleware := middleware.Func(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := user.InjectOrgID(r.Context(), "fake")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	frontend.RegisterFrontendServer(serv.GRPC, f)
	serv.HTTP.Path("/").Handler(authMiddleware.Wrap(f.Handler()))

	// start frontend
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = serv.Run()
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	// worker
	workerCfg := worker.WorkerConfig{
		Address:           "localhost:9095",
		Parallelism:       10,
		DNSLookupDuration: 10 * time.Second,
		GRPCClientConfig: grpcclient.Config{
			MaxRecvMsgSize: 10 * 1024 * 1024,
			MaxSendMsgSize: 10 * 1024 * 1024,
		},
	}
	w, err := worker.NewWorker(workerCfg, util.Logger)
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

	wg.Wait()
	cancel()
	w.Stopping(nil)
}
