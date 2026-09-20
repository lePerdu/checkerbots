#!/bin/sh

set -ex

RPI_HOST=rpi@192.168.40.107

ssh_opts='-o ControlPath=~/.ssh/master-$$ -o ControlMaster=auto -o ControlPersist=5'

cd "$(dirname "$0")"

mkdir -p bin
GOOS=linux GOARCH=arm64 go build -o bin/roomba-controller .

ssh $ssh_opts $RPI_HOST mkdir -p \~/.local/bin \~/.local/share/systemd/user
ssh $ssh_opts $RPI_HOST systemctl --user stop roomba-controller
scp $ssh_opts ./bin/roomba-controller $RPI_HOST:\~/.local/bin/roomba-controller
scp ./roomba-controller.service $RPI_HOST:\~/.local/share/systemd/user/
ssh $ssh_opts $RPI_HOST systemctl --user daemon-reload
ssh $ssh_opts $RPI_HOST systemctl --user start roomba-controller
