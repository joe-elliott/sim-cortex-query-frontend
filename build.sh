pushd frontend 
docker build . -t sim-frontend:latest
popd

pushd worker
docker build . -t sim-worker:latest
popd

pushd loadgen
docker build . -t sim-loadgen:latest
popd