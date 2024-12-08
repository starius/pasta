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

query="/api/create?"

url_shortener="no"

for o in "$@"; do
    case "$o" in

        --self-burning)
            query="${query}self_burning=yes&"
        ;;

        --long-id)
            query="${query}long_id=yes&"
        ;;

        --url-shortener)
            url_shortener="yes"
            query="${query}redirect=yes&"
        ;;

        --local)
            server="http://127.0.0.1:8042"
        ;;

        *)
            if [ $url_shortener = "yes" ]; then
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
    "${server}${query}" \
    -F "content=$content" \
    -F "filename=$filename"
