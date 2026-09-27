#!/bin/sh

set -ex

project_dir="$(dirname "$0")/../.."
go build -o "$project_dir/bin/sim-robot" "$project_dir/apps/sim-robot"

export WS_ENDPOINT=ws://localhost:8080/api/robots/ws
pids=
for i in $(seq 1 $1)
do
    "$project_dir/bin/sim-robot" r$i &
    pids="$pids $!"
done

trap 'for pid in $pids; do kill $pid; done; exit' INT TERM

for pid in $pids; do wait $pid; done
