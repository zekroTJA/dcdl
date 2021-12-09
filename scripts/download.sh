#!/bin/bash

function check {
    which $1 > /dev/null 2>&1 || {
        echo "'$1' is not installed. Please install it and re-run the script."
        exit 1
    }
}

function dl {
    echo "Download '$2'..."
    curl -sLo "$1/$2" $3
}

check jq
check curl

FILE="metadata.json"
[ -z $1 ] || FILE="$1"

[ -f $FILE ] || {
    echo "Metadata file '$FILE' could not be found."
    exit 1
}

OUTPUT="files"
[ -z $2 ] || OUTPUT="$2"

[ -d $OUTPUT ] || mkdir -p $OUTPUT

for ATT in $(cat $FILE | jq -rc '.[].attachments[] | [ .archive_filename, .url ]'); do
    dl $OUTPUT $(echo $ATT | jq -r 'join(" ")')
done