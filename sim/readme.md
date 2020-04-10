k3d create --enable-registry --name "cortex-sim"

https://github.com/rancher/k3d/blob/master/docs/registries.md#using-a-local-registry
../build.sh

kc create -f .