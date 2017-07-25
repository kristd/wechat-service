#!/usr/bin/env bash

logfile=console.log
nohup ./service -config=./cfg.json -v=2 -alsologtostderr=true -log_dir=./log &> $logfile &
