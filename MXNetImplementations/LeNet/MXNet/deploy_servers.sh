#!/bin/bash

export COMMAND='python main.py'
DMLC_ROLE=server DMLC_PS_ROOT_URI=192.168.10.205 DMLC_PS_ROOT_PORT=9092 DMLC_NUM_SERVER=1 DMLC_NUM_WORKER=1 $COMMAND &
DMLC_ROLE=scheduler DMLC_PS_ROOT_URI=192.168.10.205 DMLC_PS_ROOT_PORT=9092 DMLC_NUM_SERVER=1 DMLC_NUM_WORKER=1 $COMMAND