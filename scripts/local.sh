#!/bin/bash
go mod tidy

# Define the tag for the logger
TAG="sbcwaste"

until go run ./src 2>&1 | while IFS= read -r line; do
    echo "$line" | logger -t $TAG
done; do
    echo "$(date '+%Y-%m-%d %H:%M:%S') - Server crashed with exit code $?. Respawning.." | logger -t $TAG
    sleep 1
done

# view the output of the logs: 
#     journalctl -t sbcwaste