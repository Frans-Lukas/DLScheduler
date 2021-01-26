### Install Hadoop locally
1. Follow this guide [https://hadoop.apache.org/docs/stable/hadoop-project-dist/hadoop-common/SingleCluster.html](https://hadoop.apache.org/docs/stable/hadoop-project-dist/hadoop-common/SingleCluster.html).
2. If you get the following error when starting `start-dfs.sh`: `localhost: rcmd: socket: Permission denied`
    1. Add `export PDSH_RCMD_TYPE=ssh` to your `~/.bashrc` file.
    2. See [https://stackoverflow.com/a/48415037](https://stackoverflow.com/a/48415037)
