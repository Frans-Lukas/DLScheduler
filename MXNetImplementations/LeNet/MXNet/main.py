import os
import re
import sys
from random import randint

import mxnet
import mxnet as mx
import mxnet.autograd as ag
import mxnet.metric
import mxnet.ndarray as F
from hdfs import Config, Client
from mxnet import gluon
from mxnet.gluon import nn
from mxnet.test_utils import get_mnist_iterator

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


def main():
    # os.environ["DMLC_ROLE"] = sys.argv[1]
    print("test print in main to see if its updating on push")
    print("role: " + os.getenv("DMLC_ROLE"))
    start_lenet(Config().get_client('dev'))


def start_lenet(client: Client):
    kv = mxnet.kv.create('dist_async')
    mx.random.seed(42)
    batch_size = 100
    train_data, val_data = get_mnist_iterator(batch_size, (1, 28, 28), num_parts=kv.num_workers, part_index=kv.rank)
    net = Net()
    net = load_model(net, client)
    ctx = [mx.gpu() if mx.test_utils.list_gpus() else mx.cpu()]
    net.initialize(mx.init.Xavier(magnitude=2.24), ctx=ctx)
    trainer = gluon.Trainer(net.collect_params(), 'sgd', {'learning_rate': 0.03}, kvstore=kv)
    metric = mx.metric.Accuracy()
    softmax_cross_entropy_loss = gluon.loss.SoftmaxCrossEntropyLoss()
    epoch = 1
    loss, accuracy = train(ctx, epoch, metric, net, softmax_cross_entropy_loss, train_data, trainer)

    save_model_to_hdfs(net, client)
    print("printing works!")
    print(accuracy)
    loss = re.search('\[(.*)\]', str(loss)).group(1)
    print(
        "regexpresultstart{\"loss\":" + loss + ", \"accuracy\":" + str(accuracy) + ", \"worker_id\":0}regexpresultend")


def load_model_from_hdfs(client: Client):
    files = client.list('models')
    models = list()
    for file in files:
        client.download('models/' + file, MODEL_WEIGHTS_PATH, overwrite=True)
    # with client.read('models/' + file) as reader:
    # models.append(keras.models.load_model(MODEL_WEIGHTS_PATH))
    return models


def load_model(net: Net, client: Client):
    load_model_from_hdfs(client)
    if os.path.isfile(MODEL_WEIGHTS_PATH):
        net.load_params(MODEL_WEIGHTS_PATH)
    return net


def save_model_to_hdfs(net: Net, client: Client, node_id="0", name=""):
    net.save_params(MODEL_WEIGHTS_PATH)
    if os.path.exists(MODEL_WEIGHTS_PATH):
        if name == "":
            name = 'weights-' + str(node_id) + '.json'
        print("saving trained model to hdfs: " + name)
        try:
            with open(MODEL_WEIGHTS_PATH, 'rb') as model, client.write('models/' + name, overwrite=True) as writer:
                writer.write(model.read())
        except:
            exit(-1)


def init_context(context):
    setattr(context.user_data, HDFS_CONNECTION, Config().get_client('dev'))
    setattr(context.user_data, NODE_ID, randint(1, INT_MAX))


def start_from_nuclio(context, event):
    context.logger.info_with('Got invoked',
                             trigger_kind=event.trigger.kind,
                             event_body=event.body)
    os.environ["DMLC_ROLE"] = "worker"
    client = getattr(context.user_data, HDFS_CONNECTION)
    acc = start_lenet(client)
    return "training successful, acc: " + str(acc)


def train(ctx, epoch, metric, net, softmax_cross_entropy_loss, train_data, trainer):
    loss = any
    accuracy = any
    for i in range(epoch):
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
            # Updates internal evaluation
            metric.update(label, outputs)
            # Make one step of parameter update. Trainer needs to know the
            # batch size of data to normalize the gradient by 1/batch_size.
            trainer.step(batch.data[0].shape[0])
        # Gets the evaluation result.
        name, accuracy = metric.get()
        # Reset evaluation result to initial state.
        metric.reset()
    loss = loss.mean()
    return loss, accuracy


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
