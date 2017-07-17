#!/usr/bin/env bash

logfile=console.log
nohup ./service -v=2 -alsologtostderr=true -log_dir=../../log &> $logfile &