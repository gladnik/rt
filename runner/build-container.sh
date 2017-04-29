#!/bin/bash
set -e
set -x
if [ -z "$1" ]; then
    echo "Usage: build-container.sh <type> [<tag>]"
    exit 1
fi
type=$1
if [ ! -d "$type" ]; then
    echo "Missing directory: $type"
    exit 1
fi
tag=${2:-"latest"}
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build
cp runner "$type"
pushd "$type"
docker build -t aerokube/"$type":"$tag" .
popd