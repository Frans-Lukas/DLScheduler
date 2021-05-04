#!/bin/bash

#TODO remember that the commented out tests need to be run again



sudo /etc/kubernetes/sched-manager/enable-gang.sh
sleep 100
#./test_runner.sh 3 1
#./test_runner.sh 4 1
./test_runner.sh 17 1
./test_runner.sh 18 1
./test_runner.sh 19 1
./test_runner.sh 20 1
./test_runner.sh 21 1
./test_runner.sh 3 2
./test_runner.sh 4 2
./test_runner.sh 17 2
./test_runner.sh 18 2
./test_runner.sh 19 2
./test_runner.sh 20 2
./test_runner.sh 21 2
sudo /etc/kubernetes/sched-manager/enable-default.sh
sleep 100

./test_runner.sh 1 1
./test_runner.sh 2 1
./test_runner.sh 5 1
./test_runner.sh 6 1
./test_runner.sh 7 1
./test_runner.sh 8 1
#./test_runner.sh 9 1 3w3w
#./test_runner.sh 10 1 3w3w
./test_runner.sh 11 1
./test_runner.sh 12 1
#./test_runner.sh 13 1 3w3w
#./test_runner.sh 14 1 3w3w
./test_runner.sh 15 1
./test_runner.sh 16 1
./test_runner.sh 15 1
./test_runner.sh 16 1
./test_runner.sh 22 1
./test_runner.sh 23 1

./test_runner.sh 1 2
./test_runner.sh 2 2
./test_runner.sh 5 2
./test_runner.sh 6 2
./test_runner.sh 7 2
./test_runner.sh 8 2
#./test_runner.sh 9 1 3w3w
#./test_runner.sh 10 1 3w3w
./test_runner.sh 11 2
./test_runner.sh 12 2
#./test_runner.sh 13 1 3w3w
#./test_runner.sh 14 1 3w3w
./test_runner.sh 15 2
./test_runner.sh 16 2
./test_runner.sh 15 2
./test_runner.sh 16 2
./test_runner.sh 22 2
./test_runner.sh 23 2


#TODO uncomment these to run the Cifar10 tests
