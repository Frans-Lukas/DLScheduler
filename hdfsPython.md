### Installing a hdfs library to python
1. See documentation at [https://hdfscli.readthedocs.io/en/latest/index.html](https://hdfscli.readthedocs.io/en/latest/index.html)
2. Make sure you have a configfile `~/.hdfscli.cfg` with the following content:

```
[global]
default.alias = dev

[dev.alias]
url = http://localhost:$NAMENODE_PORT #(defaults to 9870)
user = $YOUR_USER


```
