pushd frontend 
docker build . -t frontend:latest
popd

pushd worker
docker build . -t worker:latest
popd