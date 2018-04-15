#!/usr/bin/env bash

set -eou pipefail
#set -x  # useful for debugging

docker_cleanup() {
    echo "cleaning up existing network and containers..."
    CONTAINERS='user'
    docker ps | grep -E ${CONTAINERS} | awk '{print $1}' | xargs -I {} docker stop {} || true
    docker ps -a | grep -E ${CONTAINERS} | awk '{print $1}' | xargs -I {} docker rm {} || true
    docker network list | grep ${CONTAINERS} | awk '{print $2}' | xargs -I {} docker network rm {} || true
}

# optional settings (generally defaults should be fine, but sometimes useful for debugging)
USER_LOG_LEVEL="${USER_LOG_LEVEL:-INFO}"  # or DEBUG
USER_TIMEOUT="${USER_TIMEOUT:-5}"  # 10, or 20 for really sketchy network

# local and filesystem constants
LOCAL_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# container command constants
USER_IMAGE="gcr.io/elixir-core-prod/user:snapshot" # develop

echo
echo "cleaning up from previous runs..."
docker_cleanup

echo
echo "creating user docker network..."
docker network create user

echo
echo "starting user..."
port=10100
name="user-0"
docker run --name "${name}" --net=user -d -p ${port}:${port} ${USER_IMAGE} \
    start \
    --logLevel "${USER_LOG_LEVEL}" \
    --serverPort ${port} \
    --storageMemory
user_addrs="${name}:${port}"
user_containers="${name}"

echo
echo "testing user health..."
docker run --rm --net=user ${USER_IMAGE} test health \
    --addresses "${user_addrs}" \
    --logLevel "${USER_LOG_LEVEL}"

echo
echo "testing user ..."
docker run --rm --net=user ${USER_IMAGE} test io \
    --addresses "${user_addrs}" \
    --logLevel "${USER_LOG_LEVEL}"

echo
echo "cleaning up..."
docker_cleanup

echo
echo "All tests passed."
