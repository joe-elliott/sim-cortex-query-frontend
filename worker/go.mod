module github.com/joe-elliott/sim-query-frontend

go 1.14

require (
	github.com/cortexproject/cortex v1.0.0
	github.com/go-kit/kit v0.9.0
	github.com/pkg/errors v0.9.1
	github.com/weaveworks/common v0.0.0-20200310113808-2708ba4e60a4
	google.golang.org/grpc v1.25.1
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab