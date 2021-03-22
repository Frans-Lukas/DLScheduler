import sys

import mxnet
import mxnet as mx
from mxnet import gluon
from mxnet import nd, autograd

MODEL_WEIGHTS_PATH = "/tmp/model_params.h5"

NODE_ID = 'node_id'
HDFS_CONNECTION = 'hdfs_connection'
INT_MAX = sys.maxsize


def real_fn(X):
    return 2 * X[:, 0] - 3.4 * X[:, 1] + 4.2


def main():
    # print("role: " + os.getenv("DMLC_ROLE"))
    data_ctx = mx.cpu()
    model_ctx = mx.cpu()

    num_inputs = 2
    num_outputs = 1
    num_examples = 200

    X = nd.random_normal(shape=(num_examples, num_inputs))
    noise = 0.01 * nd.random_normal(shape=(num_examples,))
    y = real_fn(X) + noise
    batch_size = 4
    kv = mxnet.kv.create('dist_sync')
    train_data = gluon.data.DataLoader(gluon.data.ArrayDataset(X, y).shard(kv.num_workers, kv.rank),
                                       batch_size=batch_size, shuffle=True)

    net = gluon.nn.Dense(1)
    net.collect_params().initialize(mx.init.Normal(sigma=1.), ctx=model_ctx)
    example_data = nd.array([[4, 7]])
    net(example_data)
    square_loss = gluon.loss.L2Loss()
    trainer = gluon.Trainer(net.collect_params(), 'sgd', {'learning_rate': 0.0001}, kvstore=kv)
    epochs = 1
    loss_sequence = []
    num_batches = num_examples / batch_size

    for e in range(epochs):
        cumulative_loss = 0
        # inner loop
        for i, (data, label) in enumerate(train_data):
            data = data.as_in_context(model_ctx)
            label = label.as_in_context(model_ctx)
            with autograd.record():
                output = net(data)
                loss = square_loss(output, label)
            loss.backward()
            trainer.step(batch_size)
            cumulative_loss += nd.mean(loss).asscalar()
        print("Epoch %s, loss: %s" % (e, cumulative_loss / num_examples))
        loss_sequence.append(cumulative_loss)
    rank = kv.rank
    print(
        "regexpresultstart{\"loss\":" + str(
            cumulative_loss / num_examples) + ", \"accuracy\":0" + ", \"worker_id\":" + str(rank) + "}regexpresultend")


if __name__ == '__main__': main()
