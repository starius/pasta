#!/bin/sh

set -ue

# Usage:
# pasta.sh file
# pasta.sh file --password
# pasta.sh file --self-burning
# cat file | pasta.sh
# cat file | pasta.sh --password
# cat file | pasta.sh --self-burning

filename=""
content="-"
pasta_type="standard"
mask='^view:'
server="https://pasta.cf"

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

        --local)
            server="http://127.0.0.1:25516"
        ;;

        *)
            filename=$(basename "$o")
            content="$o"
        ;;
    esac
done

mask="($mask)|(There was an error)"

response=$(curl \
    --silent --show-error \
    $server/api/create \
    -F "content=<$content" \
    -F "filename=$filename" \
    -F "pasta_type=$pasta_type")
echo "$response" \
| egrep "$mask" \
| sort --reverse
