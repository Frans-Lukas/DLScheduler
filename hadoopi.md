### Install Hadoop locally
1. Follow this guide [https://hadoop.apache.org/docs/stable/hadoop-project-dist/hadoop-common/SingleCluster.html](https://hadoop.apache.org/docs/stable/hadoop-project-dist/hadoop-common/SingleCluster.html).
2. If you get the following error when starting `start-dfs.sh`: `localhost: rcmd: socket: Permission denied`
    1. Add `export PDSH_RCMD_TYPE=ssh` to your `~/.bashrc` file.
    2. See [https://stackoverflow.com/a/48415037](https://stackoverflow.com/a/48415037)
    
    
## To start hdfs after installation
1. Check if it is already running with `jps`
2. If not, format namenode with `hdfs namenode -format`
3. Call `./hadoop/sbin/start-dfs.sh`
4. If something is already running, call `./hadoop/sbin/stop-all.sh` and try from step 2 again.

## The folder `user` does not exist?
For some reason hadoop does not create the correct folders. To create the user folders use:
`hadoop fs -mkdir /user`
`hadoop fs -mkdir /user/$USER`
`hadoop fs -mkdir /user/$USER/models`
