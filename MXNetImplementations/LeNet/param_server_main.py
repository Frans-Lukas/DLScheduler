import logging
import os

import mxnet
import mxnet as mx
import mxnet.autograd as ag
import mxnet.metric
import mxnet.ndarray as F
from mxnet import gluon
from mxnet.gluon import nn
from mxnet.test_utils import get_mnist_iterator


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
    print("role: " + os.getenv("DMLC_ROLE"))
    start_lenet()


def start_lenet():
    logging.basicConfig(level=logging.INFO)
    fh = logging.FileHandler('/tmp/image-classification.log')
    logger = logging.getLogger()
    logger.addHandler(fh)
    formatter = logging.Formatter('%(message)s')
    fh.setFormatter(formatter)
    fh.setLevel(logging.DEBUG)
    logging.debug('\n%s', '-' * 100)
    formatter = logging.Formatter('%(asctime)s %(levelname)s %(message)s')
    fh.setFormatter(formatter)

    logger.info('Starting new image-classification task')

    kv = mxnet.kv.create('dist_sync')
    mx.random.seed(42)
    batch_size = 100
    train_data, val_data = get_mnist_iterator(batch_size, (1, 28, 28), num_parts=kv.num_workers, part_index=kv.rank)
    # train_data = mx.io.NDArrayIter(mnist['train_data'], mnist['train_label'], batch_size, shuffle=True)
    # val_data = mx.io.NDArrayIter(mnist['test_data'], mnist['test_label'], batch_size)
    net = Net()
    ctx = [mx.gpu() if mx.test_utils.list_gpus() else mx.cpu()]
    net.initialize(mx.init.Xavier(magnitude=2.24), ctx=ctx)
    trainer = gluon.Trainer(net.collect_params(), 'sgd', {'learning_rate': 0.03}, kvstore=kv)
    metric = mx.metric.Accuracy()
    softmax_cross_entropy_loss = gluon.loss.SoftmaxCrossEntropyLoss()
    epoch = 2
    train(ctx, epoch, metric, net, softmax_cross_entropy_loss, train_data, trainer, logger)
    evaluate(ctx, net, val_data)


def start_from_nuclio(context, event):
    context.logger.info_with('Got invoked',
                             trigger_kind=event.trigger.kind,
                             event_body=event.body)
    start_lenet()
    return "training successful"


def train(ctx, epoch, metric, net, softmax_cross_entropy_loss, train_data, trainer, logger):
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
        name, acc = metric.get()
        # Reset evaluation result to initial state.
        metric.reset()
        logger.info('training acc at epoch %d: %s=%f' % (i, name, acc))


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
    assert metric.get()[1] > 0.98


if __name__ == '__main__': main()
