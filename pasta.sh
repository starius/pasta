#!/bin/sh

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
server=http://127.0.0.1:25516

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

        *)
            filename=$(basename "$o")
            content="$o"
        ;;
    esac
done

curl \
    --silent \
    $server/api/create \
    -F "content=<$content" \
    -F "filename=$filename" \
    -F "pasta_type=$pasta_type" \
| egrep "$mask" \
| sort --reverse
