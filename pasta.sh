#!/bin/sh

set -ue

# Usage:
# pasta.sh file
# pasta.sh file --password
# pasta.sh file --self-burning
# cat file | pasta.sh
# cat file | pasta.sh --password
# cat file | pasta.sh --self-burning
# pasta.sh http://example.com --url-shortener

filename=""
content="-"
pasta_type="standard"
mask='^view:'
server="https://pasta.cf"
url_shortener="no"

for o in "$@"; do
    case "$o" in

        --password)
            pasta_type="editable"
            mask="^(password|view):"
        ;;

        --self-burning)
            pasta_type="self_burning"
            mask="^(raw|view):"
        ;;

        --url-shortener)
            pasta_type="url_shortener"
            mask="^view:"
            url_shortener="yes"
        ;;

        --local)
            server="http://127.0.0.1:25516"
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

mask="($mask)|(There was an error|Failed to create paste)"

response=$(curl \
    --silent --show-error \
    $server/api/create \
    -F "content=$content" \
    -F "filename=$filename" \
    -F "pasta_type=$pasta_type")
echo "$response" \
| egrep "$mask" \
| sort --reverse
