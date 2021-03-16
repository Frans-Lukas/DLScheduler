#!/bin/bash

./invoke_mxnet.sh mxworker $1 worker &
./invoke_mxnet.sh mxserver $1 server &
./invoke_mxnet.sh mxscheduler $1 scheduler