#!/bin/bash
export COMMAND='python3 main.py'
DMLC_ROLE=server DMLC_PS_ROOT_URI=127.0.0.1 DMLC_PS_ROOT_PORT=9092 DMLC_NUM_SERVER=2 DMLC_NUM_WORKER=2 $COMMAND &
DMLC_ROLE=server DMLC_PS_ROOT_URI=127.0.0.1 DMLC_PS_ROOT_PORT=9092 DMLC_NUM_SERVER=2 DMLC_NUM_WORKER=2 $COMMAND &
DMLC_ROLE=scheduler DMLC_PS_ROOT_URI=127.0.0.1 DMLC_PS_ROOT_PORT=9092 DMLC_NUM_SERVER=2 DMLC_NUM_WORKER=2 $COMMAND &
DMLC_ROLE=worker GOOGLE_APPLICATION_CREDENTIALS=/home/franslukas/PycharmProjects/DLScheduler/MXNetImplementations/nuclio_executioner/secret_key.json DMLC_PS_ROOT_URI=127.0.0.1 DMLC_PS_ROOT_PORT=9092 DMLC_NUM_SERVER=2 DMLC_NUM_WORKER=2 $COMMAND &
DMLC_ROLE=worker GOOGLE_APPLICATION_CREDENTIALS=/home/franslukas/PycharmProjects/DLScheduler/MXNetImplementations/nuclio_executioner/secret_key.json DMLC_PS_ROOT_URI=127.0.0.1 DMLC_PS_ROOT_PORT=9092 DMLC_NUM_SERVER=2 DMLC_NUM_WORKER=2 $COMMAND
