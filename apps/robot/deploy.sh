#!/bin/sh

set -ex

cd "$(dirname "$0")"
GOOS=linux GOARCH=arm64 go build -o roomba-controller .
scp ./roomba-controller rpi@192.168.40.107:~
