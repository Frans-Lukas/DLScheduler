#!/bin/bash

#TODO remember that the commented out tests need to be run again

#sudo /etc/kubernetes/sched-manager/enable-gang.sh
#sleep 100
#./test_runner.sh 2 1
#ubernetes/sched-manager/enable-gang.sh
#sleep 100
#./test_runner.sh 21 1
#sudo /etc/kubernetes/sched-manager/enable-default.sh
#sleep 100
#./test_runner.sh 6 1

sudo /etc/kubernetes/sched-manager/enable-gang.sh
sleep 100
./test_runner.sh 24 1 6
./test_runner.sh 24 1 7
./test_runner.sh 24 1 8
./test_runner.sh 24 1 9
./test_runner.sh 24 1 10
./test_runner.sh 24 1 11
./test_runner.sh 24 1 12
./test_runner.sh 24 1 13
./test_runner.sh 24 1 14
./test_runner.sh 24 1 15
./test_runner.sh 24 1 16
./test_runner.sh 24 1 17
sudo /etc/kubernetes/sched-manager/enable-default.sh
sleep 100
./test_runner.sh 25 1 6
./test_runner.sh 25 1 7
./test_runner.sh 25 1 8
./test_runner.sh 25 1 9
./test_runner.sh 25 1 10
./test_runner.sh 25 1 11
./test_runner.sh 25 1 12
./test_runner.sh 25 1 13
./test_runner.sh 25 1 14
./test_runner.sh 25 1 15
./test_runner.sh 25 1 16
./test_runner.sh 25 1 17
#./test_runner.sh 25 1 1
#./test_runner.sh 25 1 2
#./test_runner.sh 25 1 3
#./test_runner.sh 25 1 4
#./test_runner.sh 25 1 5

#./test_runner.sh 23 1

#./test_runner.sh 1 1
#./test_runner.sh 2 1
#./test_runner.sh 5 1
#./test_runner.sh 6 1
#./test_runner.sh 7 1
#./test_runner.sh 8 1
##./test_runner.sh 9 1 3w3w
##./test_runner.sh 10 1 3w3w
#./test_runner.sh 11 1
#./test_runner.sh 12 1
##./test_runner.sh 13 1 3w3w
##./test_runner.sh 14 1 3w3w
#./test_runner.sh 15 1
#./test_runner.sh 16 1
#./test_runner.sh 15 1
#./test_runner.sh 16 1
#./test_runner.sh 22 1
#./test_runner.sh 23 1
#
#./test_runner.sh 1 2
#./test_runner.sh 2 2
#./test_runner.sh 5 2
#./test_runner.sh 6 2
#./test_runner.sh 7 2
#./test_runner.sh 8 2
##./test_runner.sh 9 1 3w3w
##./test_runner.sh 10 1 3w3w
#./test_runner.sh 11 2
#./test_runner.sh 12 2
##./test_runner.sh 13 1 3w3w
##./test_runner.sh 14 1 3w3w
#./test_runner.sh 15 2
#./test_runner.sh 16 2
#./test_runner.sh 15 2
#./test_runner.sh 16 2
#./test_runner.sh 22 2
#./test_runner.sh 23 2

#TODO uncomment these to run the Cifar10 tests
