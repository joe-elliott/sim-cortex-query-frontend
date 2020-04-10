pushd frontend 
docker build . -t registry.local:5000/sim-frontend:latest
docker push registry.local:5000/sim-frontend:latest
popd

pushd worker
docker build . -t registry.local:5000/sim-worker:latest
docker push registry.local:5000/sim-worker:latest
popd

pushd loadgen
docker build . -t registry.local:5000/sim-loadgen:latest
docker push registry.local:5000/sim-loadgen:latest
popd