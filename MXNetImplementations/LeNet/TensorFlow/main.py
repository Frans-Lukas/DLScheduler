import json
import os

import dload as dload
import keras
import numpy as np
from hdfs import Client, Config
from tensorflow.keras import datasets
from tensorflow.keras.layers import Dense, Flatten, Conv2D, AveragePooling2D
from tensorflow.keras.losses import categorical_crossentropy
from tensorflow.keras.utils import to_categorical

# Get a ResNet50 model
from distributed_lib import DistributedHelper

AVERAGED_MODEL_NAME = "averaged_lenet_tf_model.h5"

HDFS_CONNECTION = 'hdfs_connection'
NODE_ID = 'node_id'
INT_MAX = 2147483647
CIFAR_CLASSES_PATH = '/tmp/classes.pkl'
MODEL_WEIGHTS_PATH = '/tmp/lenet_tf.h5'
CIFAR_PATH = '/tmp/CIFAR-10-images'
TEST_PATH = '/tmp/CIFAR-10-images/test'
TRAIN_PATH = '/tmp/CIFAR-10-images/train'


# The main entry point for this module
def download_images():
    if not os.path.exists(CIFAR_PATH):
        print("Downloading training and test data")
        dload.save_unzip("https://github.com/YoongiKim/CIFAR-10-images/archive/master.zip", "/tmp")
        os.rename("/tmp/CIFAR-10-images-master", CIFAR_PATH)
        # os.remove("master.zip")


def init_context(context):
    setattr(context.user_data, HDFS_CONNECTION, Config().get_client('dev'))
    # if not os.path.exists(CIFAR_PATH):
    #     download_images()


def run(context, event):
    context.logger.info_with('Got invoked',
                             trigger_kind=event.trigger.kind,
                             event_body=event.body)
    body = json.loads(event.body)
    worker_id = body['worker_id']
    number_of_workers = body['max_id']
    job_type = body['job_type']

    hdfs_client = getattr(context.user_data, HDFS_CONNECTION)
    helper = DistributedHelper(hdfs_client, worker_id, number_of_workers, AVERAGED_MODEL_NAME)
    if job_type == "train":
        return train_one_epoch(helper)
    elif job_type == "average":
        helper.aggregate_weights(LeNet())
        return "averaging successful"
    else:
        return "invalid command"


def load_data(num_classes):
    (x_train, y_train), (x_test, y_test) = datasets.fashion_mnist.load_data()
    x_train = x_train[:, :, :, np.newaxis]
    x_test = x_test[:, :, :, np.newaxis]
    y_train = to_categorical(y_train, num_classes)
    y_test = to_categorical(y_test, num_classes)
    x_train = x_train.astype('float32')
    x_test = x_test.astype('float32')
    x_train /= 255
    x_test /= 255
    return x_train, y_train


def train(model, x_train, y_train):
    return model.fit(x_train, y=y_train, epochs=1)


def train_one_epoch(helper: DistributedHelper):
    print("downloading average model")
    helper.download_averaged_model()
    num_classes = 10
    print("loading data")
    x_train, y_train = load_data(num_classes)
    x_train = helper.split_data(x_train)
    y_train = helper.split_data(y_train)
    print("loading model")
    model = load_model()
    print("training")
    history = train(model, x_train, y_train)
    print("saving locally")
    model.save_weights(MODEL_WEIGHTS_PATH)
    print("saving to hdfs")
    helper.upload_model_to_hdfs(MODEL_WEIGHTS_PATH)

    loss = str(history.history['loss'][0])
    accuracy = str(history.history['accuracy'][0])
    jsonReturn = "{\"loss\":" + loss + ", \"accuracy\":" + accuracy + ", \"worker_id\":" + str(helper.worker_id) + "}"
    return jsonReturn


def LeNet(shape=(28, 28, 1), num_classes=10):
    input = keras.layers.Input(shape=shape)
    output = Conv2D(6, kernel_size=(5, 5), strides=(1, 1), activation='tanh', padding="same")(input)
    output = AveragePooling2D(pool_size=(2, 2), strides=(2, 2), padding='valid')(output)
    output = Conv2D(16, kernel_size=(5, 5), strides=(1, 1), activation='tanh', padding='valid')(output)
    output = AveragePooling2D(pool_size=(2, 2), strides=(2, 2), padding='valid')(output)
    output = Flatten()(output)
    output = Dense(120, activation='tanh')(output)
    output = Dense(84, activation='tanh')(output)
    output = Dense(num_classes, activation='softmax')(output)

    model = keras.models.Model(inputs=input, outputs=output)
    model.compile(optimizer='adam',
                  loss=categorical_crossentropy,
                  metrics=['accuracy'])
    return model


def load_model():
    if os.path.exists(AVERAGED_MODEL_NAME):
        print("using average model")
        model = averaged_model()
    else:
        print("using fresh model")
        model = LeNet()
    return model


def local_model():
    model = LeNet()
    model.load_weights(MODEL_WEIGHTS_PATH)
    return model


def averaged_model():
    model = LeNet()
    model.load_weights(AVERAGED_MODEL_NAME)
    return model


def main():
    body = "average"
    # model = LeNet()
    # model.load_weights("/tmp/weights-2.h5")
    # print(model.get_weights())
    hdfs_client = Config().get_client('dev')
    helper = DistributedHelper(hdfs_client, 0, 3, AVERAGED_MODEL_NAME)
    helper1 = DistributedHelper(hdfs_client, 1, 3, AVERAGED_MODEL_NAME)
    helper2 = DistributedHelper(hdfs_client, 2, 3, AVERAGED_MODEL_NAME)
    while True:
        print(train_one_epoch(helper1))
        print(train_one_epoch(helper))
        print(train_one_epoch(helper2))
        helper.aggregate_weights(LeNet())




# Tell python to run main method
if __name__ == '__main__': main()
