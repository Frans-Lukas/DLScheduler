#!/bin/bash
hdfs dfs -rm -r "/*"
hdfs namenode -format
$HADOOP_HOME/sbin/stop-all.sh
$HADOOP_HOME/sbin/start-dfs.sh
hadoop fs -mkdir /user
hadoop fs -mkdir /user/franslukas
hadoop fs -mkdir /user/franslukas/models
