import json
import os
import re
import sys
import timeit

import mxnet
import mxnet as mx
import mxnet.autograd as ag
import mxnet.metric
import mxnet.ndarray as F
from mxnet import gluon
from mxnet.gluon import nn

from cloudStorage import download_simple, upload_simple

MODEL_WEIGHTS_PATH = "/tmp/model_params.h5"

NODE_ID = 'node_id'
HDFS_CONNECTION = 'hdfs_connection'
INT_MAX = sys.maxsize


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


class TestData:

    def __init__(self) -> None:
        self.id = "baseline"
        self.epochs = 0
        self.time = 0.0




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


def main():
    # os.environ["DMLC_ROLE"] = sys.argv[1]

    start_lenet()


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


def start_lenet():
    start = timeit.default_timer()

    testData = TestData()
    kv = mxnet.kv.create('dist')
    mx.random.seed(42)
    batch_size = 100

    num_parts = os.getenv("NUM_PARTS")
    if num_parts is None or num_parts == 1:
        num_parts = kv.num_workers
    train_data = get_mnist_iterator_container(batch_size, (1, 28, 28), num_parts=num_parts, part_index=kv.rank)
    net = Net()
    ctx = [mx.gpu() if mx.test_utils.list_gpus() else mx.cpu()]
    net.initialize(mx.init.Xavier(magnitude=2.24), ctx=ctx)
    trainer = gluon.Trainer(net.collect_params(), 'sgd', {'learning_rate': 0.03}, kvstore=kv)
    metric = mx.metric.Accuracy()
    softmax_cross_entropy_loss = gluon.loss.SoftmaxCrossEntropyLoss()
    epoch = 1

    # WORKS DISTRIBUTED x2 workers UP until this point!
    # Which means it is the training that freezes it...

    loss, accuracy, epochs = train(ctx, epoch, metric, net, softmax_cross_entropy_loss, train_data, trainer)

    testData.epochs = epochs

    print("printing works!")
    print(accuracy)
    loss = re.search('\[(.*)\]', str(loss)).group(1)
    print(
        "regexpresultstart{\"loss\":" + loss + ", \"accuracy\":" + str(accuracy) + ", \"worker_id\":0}regexpresultend")
    stop = timeit.default_timer()

    testData.time = stop - start

    with open('results.json', 'w') as f:
        json.dump(testData.__dict__, f)


def train(ctx, epoch, metric, net, softmax_cross_entropy_loss, train_data, trainer):
    loss = any
    accuracy = any
    target_loss = 0.85
    current_loss = 1
    concurrent_count = 0
    epochs = 0
    while concurrent_count < 3:
        # Reset the train data iterator.
        train_data.reset()
        # Loop over the train data iterator.
        for batch in train_data:

            # Splits train data into multiple slices along batch_axis
            # and copy each slice into a context.
            data = gluon.utils.split_and_load(batch.data[0], ctx_list=ctx, batch_axis=0)
            # Splits train labels into multiple slices along batch_axis
            # and copy each slice into a context.
            label = gluon.utils.split_and_load(batch.label[0], ctx_list=ctx, batch_axis=0)

            # MADE IT HERE, WHICH MEANS SOMETHING BELOW CAUSES FREEZE
            outputs = []
            # Inside training scope
            with ag.record():
                for x, y in zip(data, label):
                    z = net(x)
                    # Computes softmax cross entropy loss.
                    loss = softmax_cross_entropy_loss(z, y)
                    # Backpropogate the error for one iteration.
                    loss.backward()
                    outputs.append(z)
            # MADE IT HERE, WHICH MEANS SOMETHING BELOW CAUSES FREEZE

            # Updates internal evaluation
            metric.update(label, outputs)
            # MADE IT HERE, WHICH MEANS SOMETHING BELOW CAUSES FREEZE
            # Make one step of parameter update. Trainer needs to know the
            # batch size of data to normalize the gradient by 1/batch_size.

            trainer.step(batch.data[0].shape[0])

        loss_tmp = loss.mean()
        loss_tmp = re.search('\[(.*)\]', str(loss_tmp)).group(1)
        current_loss = float(loss_tmp)
        if current_loss < target_loss:
            concurrent_count += 1
        else:
            concurrent_count = 0
        epochs += 1


            # if os.environ["DMLC_NUM_WORKER"] == "2":
            #     # print("regexpresultstart{\"loss\":0.9, \"accuracy\":0.9, \"worker_id\":0}regexpresultend")
            #     print("batch, then train_data:")
            #     print(batch)
            #     print(train_data)
            #     return [0.99, 0.99], 0.99
            # DID NOT MAKE IT HERE, WHICH MEANS SOMETHING ABOVE FREEZES WITH TWO WORKERS
        # Gets the evaluation result.
        name, accuracy = metric.get()
        # print('[Epoch %d] training: %s' % (epoch, metric_str(name, accuracy)))
        # print('[Epoch %d] acc: %f, loss: %s' % (i, accuracy, loss_tmp))
        # Reset evaluation result to initial state.
        metric.reset()
    loss = loss.mean()
    return loss, accuracy, epochs


def evaluate(ctx, net, val_data):
    # Use Accuracy as the evaluation metric.
    metric = mx.metric.Accuracy()
    # Reset the validation data iterator.
    val_data.reset()
    # Loop over the validation data iterator.
    for batch in val_data:
        # Splits validation data into multiple slices along batch_axis
        # and copy each slice into a context.
        data = gluon.utils.split_and_load(batch.data[0], ctx_list=ctx, batch_axis=0)
        # Splits validation label into multiple slices along batch_axis
        # and copy each slice into a context.
        label = gluon.utils.split_and_load(batch.label[0], ctx_list=ctx, batch_axis=0)
        outputs = []
        for x in data:
            outputs.append(net(x))
        # Updates internal evaluation
        metric.update(label, outputs)
    print('validation acc: %s=%f' % metric.get())
    return metric.get()


if __name__ == '__main__': main()

