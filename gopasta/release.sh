#!/bin/bash

set -xue

for os in linux windows darwin; do
    for arch in 386 amd64 arm64; do
        if [[ "$os/$arch" == "darwin/386" ]]; then
            continue
        fi
        GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -buildid=" -o gopasta-$os-$arch
    done
done
