k3d create --enable-registry --name "cortex-sim"

# https://github.com/rancher/k3d/blob/master/docs/registries.md#using-a-local-registry
../build.sh

kc create -f .

```
histogram_quantile(.5, sum(rate(cortex_query_frontend_queue_duration_seconds_bucket[1m])) by (le))
```