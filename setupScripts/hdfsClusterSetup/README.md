1.
get IPs for each node
check that JDK is installed


Configure the System
create Host File on each Node
File: /etc/hosts
IP1    node-master
IP2    node1
IP3    node2

	Distribute Authentication Key-pairs for the Hadoop User
		1. Login to node-master as the hadoop user, and generate an SSH key:
			ssh-keygen -b 4096
		2. View the node-master public key and copy it to your clipboard to use with each of your worker nodes.
			less /home/hadoop/.ssh/id_rsa.pub
		3. In each node, make a new file master.pub in the /home/hadoop/.ssh directory. Paste your public key into this file and save your changes.
		4. Copy your key file into the authorized key store. (assume this is in each node)
			cat ~/.ssh/master.pub >> ~/.ssh/authorized_keys
	
	Download and Unpack Hadoop Binaries
		Log into node-master as the hadoop user, download the Hadoop tarball from Hadoop project page, and unzip it:
			cd
			wget http://apache.cs.utah.edu/hadoop/common/current/hadoop-3.1.2.tar.gz
			tar -xzf hadoop-3.1.2.tar.gz
			mv hadoop-3.1.2 hadoop
		
	Set Environment Variables
		1. Add Hadoop binaries to your PATH. Edit /home/hadoop/.profile and add the following line:
			File: /home/hadoop/.profile
				PATH=/home/hadoop/hadoop/bin:/home/hadoop/hadoop/sbin:$PATH
		2. Add Hadoop to your PATH for the shell. Edit .bashrc and add the following lines:
			File: /home/hadoop/.bashrc
				export HADOOP_HOME=/home/hadoop/hadoop
				export PATH=${PATH}:${HADOOP_HOME}/bin:${HADOOP_HOME}/sbin

Configure the Master Node
Configuration will be performed on node-master and replicated to other nodes.

	Set JAVA_HOME
		1. Find your Java installation path. This is known as JAVA_HOME. If you installed open-jdk from your package manager, you can find the path with the command:
			update-alternatives --display java
		   	
		   	Take the value of the current link and remove the trailing /bin/java. For example on Debian, the link is /usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java, so JAVA_HOME should be /usr/lib/jvm/java-8-openjdk-amd64/jre.
		   	
		   	If you installed java from Oracle, JAVA_HOME is the path where you unzipped the java archive.
		2. Edit ~/hadoop/etc/hadoop/hadoop-env.sh and replace this line:
			export JAVA_HOME=${JAVA_HOME}
			with your actual java installation path. On a Debian 9 Linode with open-jdk-8 this will be as follows:
				File: ~/hadoop/etc/hadoop/hadoop-env.sh
					export JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64/jre
	
	Set NameNode Location
		Update your ~/hadoop/etc/hadoop/core-site.xml file to set the NameNode location to node-master on port 9000:
			File: ~/hadoop/etc/hadoop/core-site.xml
				<?xml version="1.0" encoding="UTF-8"?>
				<?xml-stylesheet type="text/xsl" href="configuration.xsl"?>
				<configuration>
					<property>
            					<name>fs.default.name</name>
            					<value>hdfs://node-master:9000</value>
        				</property>
    				</configuration>
	
	Set path for HDFS
		Edit hdfs-site.conf to resemble the following configuration:
			File: ~/hadoop/etc/hadoop/hdfs-site.xml
				<configuration>
    					<property>
        					<name>dfs.namenode.name.dir</name>
            					<value>/home/hadoop/data/nameNode</value>
    					</property>

    					<property>
            					<name>dfs.datanode.data.dir</name>
            					<value>/home/hadoop/data/dataNode</value>
    					</property>

    					<property>
            					<name>dfs.replication</name>
            					<value>1</value>
    					</property>
				</configuration>
			
			The last property, dfs.replication, indicates how many times data is replicated in the cluster. You can set 3 to have all the data duplicated on the three nodes. Don’t enter a value higher than the actual number of worker nodes.
		
	Set YARN as Job Scheduler
		Edit the mapred-site.xml file, setting YARN as the default framework for MapReduce operations:
			File: ~/hadoop/etc/hadoop/mapred-site.xml
				<configuration>
					<property>
            					<name>mapreduce.framework.name</name>
            					<value>yarn</value>
    					</property>
    					<property>
            					<name>yarn.app.mapreduce.am.env</name>
            					<value>HADOOP_MAPRED_HOME=$HADOOP_HOME</value>
    					</property>
    					<property>
            					<name>mapreduce.map.env</name>
            					<value>HADOOP_MAPRED_HOME=$HADOOP_HOME</value>
    					</property>
    					<property>
            					<name>mapreduce.reduce.env</name>
            					<value>HADOOP_MAPRED_HOME=$HADOOP_HOME</value>
    					</property>
				</configuration>
	
	Configure YARN
		Edit yarn-site.xml, which contains the configuration options for YARN. In the value field for the yarn.resourcemanager.hostname, replace 203.0.113.0 with the public IP address of node-master:
			File: ~/hadoop/etc/hadoop/yarn-site.xml
				<configuration>
    					<property>
           			 		<name>yarn.acl.enable</name>
            					<value>0</value>
    					</property>

    					<property>
            					<name>yarn.resourcemanager.hostname</name>
            					<value>203.0.113.0</value>
    					</property>

    					<property>
            					<name>yarn.nodemanager.aux-services</name>
            					<value>mapreduce_shuffle</value>
    					</property>
				</configuration>
	
	Configure Workers
		The file workers is used by startup scripts to start required daemons on all nodes. Edit ~/hadoop/etc/hadoop/workers to include all of the nodes:
			File: ~/hadoop/etc/hadoop/workers
				node-master
				node1
				node2

Configure Memory Allocation
Memory allocation can be tricky on low RAM nodes because default values are not suitable for nodes with less than 8GB of RAM. This section will highlight how memory allocation works for MapReduce jobs, and provide a sample configuration for 2GB RAM nodes.

	Sample Configuration for 2GB Nodes
		For 2GB nodes, a working configuration may be:
			roperty	Value
			yarn.nodemanager.resource.memory-mb	1536
			yarn.scheduler.maximum-allocation-mb	1536
			yarn.scheduler.minimum-allocation-mb	128
			yarn.app.mapreduce.am.resource.mb	512
			mapreduce.map.memory.mb	256
			mapreduce.reduce.memory.mb	256
		
		1. Edit /home/hadoop/hadoop/etc/hadoop/yarn-site.xml and add the following lines:
			File: ~/hadoop/etc/hadoop/yarn-site.xml
				<property>
        				<name>yarn.nodemanager.resource.memory-mb</name>
        				<value>1536</value>
				</property>

				<property>
        				<name>yarn.scheduler.maximum-allocation-mb</name>
        				<value>1536</value>
				</property>

				<property>
        				<name>yarn.scheduler.minimum-allocation-mb</name>
        				<value>128</value>
				</property>

				<property>
        				<name>yarn.nodemanager.vmem-check-enabled</name>
        				<value>false</value>
				</property>
			
			The last property disables virtual-memory checking which can prevent containers from being allocated properly with JDK8 if enabled.
		
		2. Edit /home/hadoop/hadoop/etc/hadoop/mapred-site.xml and add the following lines:
			File: ~/hadoop/etc/hadoop/mapred-site.xml
				<property>
        				<name>yarn.app.mapreduce.am.resource.mb</name>
        				<value>512</value>
				</property>

				<property>
        				<name>mapreduce.map.memory.mb</name>
        				<value>256</value>
				</property>

				<property>
        				<name>mapreduce.reduce.memory.mb</name>
        				<value>256</value>
				</property>

Duplicate Config Files on Each Node
1. Copy the Hadoop binaries to the other worker nodes:
cd /home/hadoop/
scp hadoop-*.tar.gz node1:/home/hadoop
scp hadoop-*.tar.gz node2:/home/hadoop
2. Connect to node1 via SSH. A password isn’t required, thanks to the SSH keys copied above:
ssh node1
3. Unzip the binaries, rename the directory, and exit node1 to get back on the node-master:
tar -xzf hadoop-3.1.2.tar.gz
mv hadoop-3.1.2 hadoop
exit
4. Repeat steps 2 and 3 for node2.

	5. Copy the Hadoop configuration files to the worker nodes:
		for node in node1 node2; do
    			scp ~/hadoop/etc/hadoop/* $node:/home/hadoop/hadoop/etc/hadoop/;
		done

Format HDFS
HDFS needs to be formatted like any classical file system. On node-master, run the following command:
hdfs namenode -format
Your Hadoop installation is now configured and ready to run.

Run and monitor HDFS
Start and Stop HDFS
1. Start the HDFS by running the following script from node-master:
start-dfs.sh
2. check that every process is running with the jps command on each node. On node-master, you should see the following (the PID number will be different):
21922 Jps
21603 NameNode
21787 SecondaryNameNode
19728 DataNode

			And on node1 and node2 you should see the following:
			
			19728 DataNode
			19819 Jps
		
		3. To stop HDFS on master and worker nodes, run the following command from node-master:
			stop-dfs.sh
		