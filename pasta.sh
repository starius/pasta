#!/bin/sh

filename=/dev/stdin
if [ ! -z "$1" ]; then
    filename=$1
fi

server=http://127.0.0.1:25516

pasta_type="standard"

curl \
    --silent \
    $server/api/create \
    -F "content=<$filename" \
    -F "filename=$filename" \
    -F "pasta_type=$pasta_type" \
| egrep '^(raw|view):' \
| sort --reverse
