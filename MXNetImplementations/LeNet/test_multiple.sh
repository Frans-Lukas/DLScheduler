#!/bin/bash


while [ true ]
do
  ./invoke_tf_model.sh lenet2-1 0 3 train
  ./invoke_tf_model.sh lenet2-2 1 3 train
  ./invoke_tf_model.sh lenet2-3 2 3 train
  ./invoke_tf_model.sh real-para-server2 1 3 average
done