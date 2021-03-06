import logging
import os
import re
import sys
import time

import mxnet
import mxnet as mx
from mxnet import autograd as ag
from mxnet import gluon
from mxnet.gluon.model_zoo import vision
from mxnet.gluon.model_zoo.vision import SqueezeNet
from mxnet.gluon.utils import download
from mxnet.image import color_normalize

from cloudStorage import download_simple, upload_simple

MODEL_WEIGHTS_PATH = "/tmp/model_params.h5"

NODE_ID = 'node_id'
HDFS_CONNECTION = 'hdfs_connection'
INT_MAX = sys.maxsize


def real_fn(X):
    return 2 * X[:, 0] - 3.4 * X[:, 1] + 4.2


def save_model_to_hdfs(net: SqueezeNet):
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


def load_model(net: SqueezeNet):
    load_model_from_gcloud()
    if os.path.isfile(MODEL_WEIGHTS_PATH):
        net.load_params(MODEL_WEIGHTS_PATH)
    return net


def main():
    kv = mxnet.kv.create('dist')
    batch_size = 256
    log_interval = 100
    mode = 'hybrid'
    epochs = 1
    learning_rate = 0.05
    wd = 0.002
    positive_class_weight = 5
    gpus = mx.test_utils.list_gpus()
    contexts = [mx.gpu(i) for i in gpus] if len(gpus) > 0 else [mx.cpu()]

    dataset_key = 'validation'
    dataset_files = {'train': ('not_hotdog_train-e6ef27b4.rec', '0aad7e1f16f5fb109b719a414a867bbee6ef27b4'),
                     'validation': ('not_hotdog_validation-c0201740.rec', '723ae5f8a433ed2e2bf729baec6b878ac0201740')}
    num_images = {'validation': 1259, 'train': 16882}
    training_dataset_name, training_data_hash = dataset_files[dataset_key]

    def verified(file_path, sha1hash):
        import hashlib
        sha1 = hashlib.sha1()
        with open(file_path, 'rb') as f:
            while True:
                data = f.read(1048576)
                if not data:
                    break
                sha1.update(data)
        matched = sha1.hexdigest() == sha1hash
        if not matched:
            logging.warn('Found hash mismatch in file {}, possibly due to incomplete download.'
                         .format(file_path))
        return matched

    training_dataset_path = training_dataset_name
    url_format = 'https://apache-mxnet.s3-accelerate.amazonaws.com/gluon/dataset/{}'
    if not os.path.exists(training_dataset_path) or not verified(training_dataset_path, training_data_hash):
        logging.info('Downloading training dataset.')
        download(url_format.format(training_dataset_name), path=training_dataset_path, overwrite=True)

    num_parts = os.getenv("NUM_PARTS")
    if num_parts is None or num_parts == 1:
        num_parts = kv.num_workers

    batch_size = min(batch_size, int(num_images[dataset_key] / int(num_parts)))
    train_iter = mx.io.ImageRecordIter(path_imgrec=training_dataset_path,
                                       min_img_size=256,
                                       data_shape=(3, 224, 224),
                                       rand_crop=True,
                                       shuffle=True,
                                       batch_size=batch_size,
                                       max_random_scale=1.5,
                                       min_random_scale=0.75,
                                       rand_mirror=True,
                                       num_parts=num_parts,
                                       part_index=kv.rank)

    deep_dog_net = vision.squeezenet1_1(prefix='deep_dog_', classes=2)
    deep_dog_net.collect_params().initialize(ctx=contexts)
    # load_model(deep_dog_net)

    def metric_str(names, accs):
        return ', '.join(['%s=%f' % (name, acc) for name, acc in zip(names, accs)])

    metric = mx.metric.create(['acc', 'f1'])

    def train(net, train_iter, epochs, ctx):
        if isinstance(ctx, mx.Context):
            ctx = [ctx]
        trainer = gluon.Trainer(net.collect_params(), 'sgd', {'learning_rate': learning_rate, 'wd': wd}, kvstore=kv)
        loss = gluon.loss.SoftmaxCrossEntropyLoss()
        latestLoss = 0.0
        accs = None
        for epoch in range(epochs):
            tic = time.time()
            train_iter.reset()
            btic = time.time()
            for i, batch in enumerate(train_iter):
                # the model zoo models expect normalized images
                data = color_normalize(batch.data[0] / 255,
                                       mean=mx.nd.array([0.485, 0.456, 0.406]).reshape((1, 3, 1, 1)),
                                       std=mx.nd.array([0.229, 0.224, 0.225]).reshape((1, 3, 1, 1)))
                data = gluon.utils.split_and_load(data, ctx_list=ctx, batch_axis=0)
                label = gluon.utils.split_and_load(batch.label[0], ctx_list=ctx, batch_axis=0)
                outputs = []
                Ls = []
                with ag.record():
                    for x, y in zip(data, label):
                        z = net(x)
                        # rescale the loss based on class to counter the imbalance problem
                        L = loss(z, y) * (1 + y * positive_class_weight) / positive_class_weight
                        latestLoss = L
                        # store the loss and do backward after we have done forward
                        # on all GPUs for better speed on multiple GPUs.
                        Ls.append(L)
                        outputs.append(z)
                    for L in Ls:
                        L.backward()
                trainer.step(batch.data[0].shape[0])
                metric.update(label, outputs)
                if log_interval and not (i + 1) % log_interval:
                    names, accs = metric.get()
                    print('[Epoch %d Batch %d] speed: %f samples/s, training: %s' % (
                        epoch, i, batch_size / (time.time() - btic), metric_str(names, accs)))
                btic = time.time()

            names, accs = metric.get()
            metric.reset()
            print('[Epoch %d] training: %s' % (epoch, metric_str(names, accs)))
            print('[Epoch %d] time cost: %f' % (epoch, time.time() - tic))
        return latestLoss, accs[0]

    if mode == 'hybrid':
        deep_dog_net.hybridize()

    deep_dog_net.collect_params().reset_ctx(contexts)
    loss, acc = train(deep_dog_net, train_iter, epochs, contexts)

    loss = loss.mean()
    loss = re.search('\[(.*)\]', str(loss)).group(1)

    print("regexpresultstart{\"loss\":" + str(
        loss) + ", \"accuracy\":" + str(acc) + ", \"worker_id\":" + str(kv.rank) + "}regexpresultend")
    # save_model_to_hdfs(deep_dog_net)


if __name__ == '__main__': main()
