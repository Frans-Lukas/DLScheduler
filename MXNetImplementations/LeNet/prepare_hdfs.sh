#!/bin/bash
if [ $# -eq 0 ]; then
  echo "usage: ./prepare_hdfs.sh \$USER"
fi

hdfs namenode -format
/usr/local/hadoop/sbin/stop-all.sh
/usr/local/hadoop/sbin/start-dfs.sh
hadoop fs -mkdir /user
hadoop fs -mkdir /user/$1
hadoop fs -mkdir /user/$1/models
