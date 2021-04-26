#!/usr/bin/env python

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

"""cifar10_dist.py contains code that trains a ResNet18 network using distributed training"""

from __future__ import print_function

import os
import random
import re

import mxnet as mx
import numpy as np
from mxnet import autograd, gluon, kv, nd
from mxnet.gluon.model_zoo import vision
from mxnet.gluon.model_zoo.vision import ResNetV1
from mxnet.gluon import nn
import mxnet.ndarray as F

from cloudStorage import download_simple, upload_simple

MODEL_WEIGHTS_PATH = "/tmp/model_params.h5"


class Net(gluon.Block):
    def __init__(self, **kwargs):
        super(Net, self).__init__(**kwargs)
        with self.name_scope():
            # layers created in name_scope will inherit name space
            # from parent layer.
            self.conv1 = nn.Conv2D(20, kernel_size=(5, 5))
            self.pool1 = nn.MaxPool2D(pool_size=(2, 2), strides=(2, 2))
            self.conv2 = nn.Conv2D(50, kernel_size=(5, 5))
            self.pool2 = nn.MaxPool2D(pool_size=(2, 2), strides=(2, 2))
            self.fc1 = nn.Dense(500)
            self.fc2 = nn.Dense(10)

    def forward(self, x):
        x = self.pool1(F.tanh(self.conv1(x)))
        x = self.pool2(F.tanh(self.conv2(x)))
        # 0 means copy over size from corresponding dimension.
        # -1 means infer size from the rest of dimensions.
        x = x.reshape((0, -1))
        x = F.tanh(self.fc1(x))
        x = F.tanh(self.fc2(x))
        return x

def save_model_to_gcloud(net: Net):
    net.save_params(MODEL_WEIGHTS_PATH)
    name = get_weights_file_name()
    if os.path.exists(MODEL_WEIGHTS_PATH):
        print("saving trained model to hdfs: " + name)
        upload_simple(MODEL_WEIGHTS_PATH, name)


def load_model_from_gcloud():
    file_name = get_weights_file_name()
    download_simple(file_name, MODEL_WEIGHTS_PATH)


def get_weights_file_name():
    job_id = os.getenv("JOB_ID")
    file_name = job_id + ".h5"
    return file_name


def load_model(net: Net):
    load_model_from_gcloud()
    if os.path.isfile(MODEL_WEIGHTS_PATH):
        net.load_params(MODEL_WEIGHTS_PATH)
    return net


def get_mnist_iterator_container(batch_size, input_shape, num_parts=1, part_index=0):
    """Returns training and validation iterators for MNIST dataset
    """
    flat = len(input_shape) != 3

    train_dataiter = mx.io.MNISTIter(
        image="/opt/nuclio/data/train-images-idx3-ubyte",
        label="/opt/nuclio/data/train-labels-idx1-ubyte",
        input_shape=input_shape,
        batch_size=batch_size,
        shuffle=True,
        flat=flat,
        num_parts=num_parts,
        part_index=part_index)

    return train_dataiter


def main():
    # Create a distributed key-value store
    store = kv.create('dist')

    # Clasify the images into one of the 10 digits
    num_outputs = 10

    # 64 images in a batch
    batch_size_per_gpu = 64
    # How many epochs to run the training
    epochs = 10

    # How many GPUs per machine
    gpus_per_machine = 1
    # Effective batch size across all GPUs
    batch_size = batch_size_per_gpu * gpus_per_machine

    # Create the context (a list of all GPUs to be used for training)
    ctx = [mx.gpu() if mx.test_utils.list_gpus() else mx.cpu()]

    num_parts = os.getenv("NUM_PARTS")
    if num_parts is None or num_parts == 1:
        num_parts = store.num_workers
    else:
        num_parts = int(num_parts)
    # Load the training data
    train_data = get_mnist_iterator_container(batch_size, (1, 28, 28), num_parts=num_parts, part_index=store.rank)

    # Use ResNet from model zoo
    net = Net()

    # Initialize the parameters with Xavier initializer
    net.collect_params().initialize(mx.init.Xavier(), ctx=ctx)
    load_model(net)

    # Use Adam optimizer. Ask trainer to use the distributor kv store.
    trainer = gluon.Trainer(net.collect_params(), 'sgd', {'learning_rate': .001}, kvstore=store)

    # We'll use cross entropy loss since we are doing multiclass classification
    loss = gluon.loss.SoftmaxCrossEntropyLoss()

    # Run one forward and backward pass on multiple GPUs
    def forward_backward(network, data, label):

        # Ask autograd to remember the forward pass
        with autograd.record():
            # Compute the loss on all GPUs
            losses = []
            outputs = []
            for X, Y in zip(data, label):
                z = network(X)
                losses.append(loss(z, Y))
                outputs.append(z)
                # losses = [loss(network(X), Y) ]

        # metric.update(label, outputs)
        # Run the backward pass (calculate gradients) on all GPUs
        for l in losses:
            l.backward()
        return losses[0].mean()

    # Train a batch using multiple GPUs
    def train_batch(batch_list, context, network, gluon_trainer):
        # Split and load data into multiple GPUs
        data = gluon.utils.split_and_load(batch_list.data[0], ctx_list=context, batch_axis=0)

        # Split and load label into multiple GPUs
        label = gluon.utils.split_and_load(batch_list.label[0], ctx_list=context, batch_axis=0)

        # Run the forward and backward pass
        loss = forward_backward(network, data, label)

        # Update the parameters
        # print(batch_list[0].shape[0])
        gluon_trainer.step(batch_list.data[0].shape[0])
        return loss

    loss_val = None
    # Run as many epochs as required
    for epoch in range(epochs):

        # Iterate through batches and run training using multiple GPUs
        batch_num = 1
        for batch in train_data:
            # Train the batch using multiple GPUs
            loss_val = train_batch(batch, ctx, net, trainer)

            batch_num += 1

        # Print test accuracy after every epoch
        # test_accuracy = evaluate_accuracy(test_data, net)
        # print("Epoch %d: Test_acc %f" % (epoch, test_accuracy))
        # sys.stdout.flush()
    loss_val = re.search('\[(.*)\]', str(loss_val)).group(1)
    # name, acc = metric.get()
    print(
        "regexpresultstart{\"loss\":" + loss_val + ", \"accuracy\":0, \"worker_id\":0}regexpresultend")
    save_model_to_gcloud(net)


if __name__ == '__main__': main()
