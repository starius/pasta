#!/bin/sh

set -ue

server="https://pastacity.nl"

# Usage:
# $ pasta.sh file [options]
# or
# $ cat file | pasta.sh [options]
#
# Options:
#
# --self-burning
#    Generate one-time link, which is destroyed when accessed.
#
# --long-id
#    Generate long ID instead of short human-readable ID
#
# --url-shortener
#    The content is used as URL for redirect.

filename=""
content="<-"

self_burning=""
long_id=""
url_shortener=""

for o in "$@"; do
    case "$o" in

        --self-burning)
            self_burning="yes"
        ;;

        --long-id)
            long_id="yes"
        ;;

        --url-shortener)
            url_shortener="yes"
        ;;

        --local)
            server="http://127.0.0.1:8042"
        ;;

        *)
            if [ "$url_shortener" = "yes" ]; then
                content="$o"
            else
                filename=$(basename "$o")
                content="<$o"
            fi
        ;;
    esac
done

curl \
    --silent --show-error \
    "${server}/api/create" \
    -F "content=$content" \
    --form-string "filename=$filename" \
    --form-string "self_burning=$self_burning" \
    --form-string "long_id=$long_id" \
    --form-string "redirect=$url_shortener"
