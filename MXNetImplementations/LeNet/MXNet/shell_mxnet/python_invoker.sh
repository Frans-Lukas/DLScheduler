#!/bin/sh

# @nuclio.configure
#
# function.yaml:
#   apiVersion: "nuclio.io/v1"
#   kind: "NuclioFunction"
#   spec:
#     runtime: "shell"
#     handler: "python_invoker.sh"
#     env:
#       - name: DMLC_PS_ROOT_URI
#         value: 127.0.0.1
#         #value: 192.168.10.205
#       - name: DMLC_PS_ROOT_PORT
#         value: 9092
#       - name: PS_VERBOSE
#         value: 1
#       - name: DMLC_NUM_SERVER
#         value: 1
#       - name: DMLC_NUM_WORKER
#         value: 1
#       - name: DMLC_ROLE
#         value: worker
#       - name: HDFSCLI_CONFIG
#         value: /tmp/.hdfscli.cfg
#       - name: HADOOP_USER_NAME
#         value: franslukas
#     build:
#       commands:
#         - echo "[global]" >> /tmp/.hdfscli.cfg
#         - echo "default.alias = dev" >> /tmp/.hdfscli.cfg
#         - echo "[dev.alias]" >> /tmp/.hdfscli.cfg
#         - echo "url = http://192.168.10.205:9870" >> /tmp/.hdfscli.cfg
#         - echo "user = franslukas" >> /tmp/.hdfscli.cfg
#         - apk --update add --no-cache git
#         - apk --update add --no-cache --virtual .build-deps g++ python3-dev libffi-dev openssl-dev
#         - apk --update add --no-cache python3 py3-pip
#         - if [ ! -e /usr/bin/pip ]; then ln -s pip3 /usr/bin/pip ; fi
#         - if [[ ! -e /usr/bin/python ]]; then ln -sf /usr/bin/python3 /usr/bin/python; fi
#         - pip install -U pip setuptools
#         - git clone https://github.com/Frans-Lukas/DLScheduler.git


## "pip install mxnet pillow keras dload numpy hdfs"


rev /dev/stdin