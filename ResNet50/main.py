import json
import os
import pickle
from random import randint

import dload as dload
import keras
import numpy as np
from hdfs import Client, Config, HdfsError

# Get a ResNet50 model
from tensorflow.python.keras.callbacks import History

AVERAGED_MODEL_NAME = "averaged_model.h5"

HDFS_CONNECTION = 'hdfs_connection'
NODE_ID = 'node_id'
INT_MAX = 2147483647
CIFAR_CLASSES_PATH = '/tmp/classes.pkl'
MODEL_WEIGHTS_PATH = '/tmp/resnet_50.h5'
CIFAR_PATH = '/tmp/CIFAR-10-images'
TEST_PATH = '/tmp/CIFAR-10-images/test'
TRAIN_PATH = '/tmp/CIFAR-10-images/train'


# Thanks to @annytab https://www.annytab.com/
# https://www.annytab.com/resnet50-image-classification-in-python/

def resnet50_model(classes=1000, *args, **kwargs):
    # Load a model if we have saved one
    if os.path.isfile(AVERAGED_MODEL_NAME):
        return keras.models.load_model(AVERAGED_MODEL_NAME)
    if os.path.isfile(MODEL_WEIGHTS_PATH):
        return keras.models.load_model(MODEL_WEIGHTS_PATH)
    # Create an input layer 
    input = keras.layers.Input(shape=(None, None, 3))
    # Create output layers
    output = keras.layers.ZeroPadding2D(padding=3, name='padding_conv1')(input)
    output = keras.layers.Conv2D(64, (7, 7), strides=(2, 2), use_bias=False, name='conv1')(output)
    output = keras.layers.BatchNormalization(axis=3, epsilon=1e-5, name='bn_conv1')(output)
    output = keras.layers.Activation('relu', name='conv1_relu')(output)
    output = keras.layers.MaxPooling2D((3, 3), strides=(2, 2), padding='same', name='pool1')(output)
    output = conv_block(output, 3, [64, 64, 256], stage=2, block='a', strides=(1, 1))
    output = identity_block(output, 3, [64, 64, 256], stage=2, block='b')
    output = identity_block(output, 3, [64, 64, 256], stage=2, block='c')
    output = conv_block(output, 3, [128, 128, 512], stage=3, block='a')
    output = identity_block(output, 3, [128, 128, 512], stage=3, block='b')
    output = identity_block(output, 3, [128, 128, 512], stage=3, block='c')
    output = identity_block(output, 3, [128, 128, 512], stage=3, block='d')
    output = conv_block(output, 3, [256, 256, 1024], stage=4, block='a')
    output = identity_block(output, 3, [256, 256, 1024], stage=4, block='b')
    output = identity_block(output, 3, [256, 256, 1024], stage=4, block='c')
    output = identity_block(output, 3, [256, 256, 1024], stage=4, block='d')
    output = identity_block(output, 3, [256, 256, 1024], stage=4, block='e')
    output = identity_block(output, 3, [256, 256, 1024], stage=4, block='f')
    output = conv_block(output, 3, [512, 512, 2048], stage=5, block='a')
    output = identity_block(output, 3, [512, 512, 2048], stage=5, block='b')
    output = identity_block(output, 3, [512, 512, 2048], stage=5, block='c')
    output = keras.layers.GlobalAveragePooling2D(name='pool5')(output)
    output = keras.layers.Dense(classes, activation='softmax', name='fc1000')(output)
    # Create a model from input layer and output layers
    model = keras.models.Model(inputs=input, outputs=output, *args, **kwargs)
    # Print model
    print()
    print(model.summary(), '\n')
    # Compile the model
    model.compile(loss='categorical_crossentropy', optimizer=keras.optimizers.Adam(lr=0.01, clipnorm=0.001),
                  metrics=['accuracy'])
    # Return a model
    return model


# Create an identity block
def identity_block(input, kernel_size, filters, stage, block):
    # Variables
    filters1, filters2, filters3 = filters
    conv_name_base = 'res' + str(stage) + block + '_branch'
    bn_name_base = 'bn' + str(stage) + block + '_branch'
    # Create layers
    output = keras.layers.Conv2D(filters1, (1, 1), kernel_initializer='he_normal', name=conv_name_base + '2a')(input)
    output = keras.layers.BatchNormalization(axis=3, name=bn_name_base + '2a')(output)
    output = keras.layers.Activation('relu')(output)
    output = keras.layers.Conv2D(filters2, kernel_size, padding='same', kernel_initializer='he_normal',
                                 name=conv_name_base + '2b')(output)
    output = keras.layers.BatchNormalization(axis=3, name=bn_name_base + '2b')(output)
    output = keras.layers.Conv2D(filters3, (1, 1), kernel_initializer='he_normal', name=conv_name_base + '2c')(output)
    output = keras.layers.BatchNormalization(axis=3, name=bn_name_base + '2c')(output)
    output = keras.layers.add([output, input])
    output = keras.layers.Activation('relu')(output)
    # Return a block
    return output


# Create a convolution block
def conv_block(input, kernel_size, filters, stage, block, strides=(2, 2)):
    # Variables
    filters1, filters2, filters3 = filters
    conv_name_base = 'res' + str(stage) + block + '_branch'
    bn_name_base = 'bn' + str(stage) + block + '_branch'
    # Create block layers
    output = keras.layers.Conv2D(filters1, (1, 1), strides=strides, kernel_initializer='he_normal',
                                 name=conv_name_base + '2a')(input)
    output = keras.layers.BatchNormalization(axis=3, name=bn_name_base + '2a')(output)
    output = keras.layers.Activation('relu')(output)
    output = keras.layers.Conv2D(filters2, kernel_size, padding='same', kernel_initializer='he_normal',
                                 name=conv_name_base + '2b')(output)
    output = keras.layers.BatchNormalization(axis=3, name=bn_name_base + '2b')(output)
    output = keras.layers.Activation('relu')(output)
    output = keras.layers.Conv2D(filters3, (1, 1), kernel_initializer='he_normal', name=conv_name_base + '2c')(output)
    output = keras.layers.BatchNormalization(axis=3, name=bn_name_base + '2c')(output)
    shortcut = keras.layers.Conv2D(filters3, (1, 1), strides=strides, kernel_initializer='he_normal',
                                   name=conv_name_base + '1')(input)
    shortcut = keras.layers.BatchNormalization(axis=3, name=bn_name_base + '1')(shortcut)
    output = keras.layers.add([output, shortcut])
    output = keras.layers.Activation('relu')(output)
    # Return a block
    return output


# Train a model
def train(steps_per_epoch, batch_size):
    # Variables, 25 epochs so far
    epochs = 1
    batch_size = 32
    train_samples = 10 * 5000  # 10 categories with 5000 images in each category
    validation_samples = 10 * 1000  # 10 categories with 1000 images in each category
    img_width, img_height = 32, 32
    # Get the model (10 categories)
    model = resnet50_model(10)
    # Create a data generator for training
    train_data_generator = keras.preprocessing.image.ImageDataGenerator(
        rescale=1. / 255,
        shear_range=0.2,
        zoom_range=0.2,
        horizontal_flip=True)
    # Create a data generator for validation
    validation_data_generator = keras.preprocessing.image.ImageDataGenerator(
        rescale=1. / 255,
        shear_range=0.2,
        zoom_range=0.2,
        horizontal_flip=True)
    # Create a train generator
    train_generator = train_data_generator.flow_from_directory(
        TRAIN_PATH,
        target_size=(img_width, img_height),
        batch_size=batch_size,
        color_mode='rgb',
        shuffle=True,
        class_mode='categorical')
    # Create a test generator
    # validation_generator = validation_data_generator.flow_from_directory(
    #     TEST_PATH,
    #     target_size=(img_width, img_height),
    #     batch_size=batch_size,
    #     color_mode='rgb',
    #     shuffle=True,
    #     class_mode='categorical')
    # Start training, fit the model
    history = model.fit_generator(
        train_generator,
        steps_per_epoch=steps_per_epoch,
        # validation_data=validation_generator,
        # validation_steps=validation_samples // batch_size,
        epochs=epochs)
    # Save model to disk
    print(model.weights)
    model.save(MODEL_WEIGHTS_PATH)
    print('Saved model to disk!')
    # Get labels
    # labels = train_generator.class_indices
    # Invert labels
    # classes = {}
    # for key, value in labels.items():
    #     classes[value] = key.capitalize()
    # Save classes to file
    # with open(CIFAR_CLASSES_PATH, 'wb') as file:
    #     pickle.dump(classes, file)
    # print('Saved classes to disk!')
    return history


# The main entry point for this module
def download_images():
    if not os.path.exists(CIFAR_PATH):
        print("Downloading training and test data")
        dload.save_unzip("https://github.com/YoongiKim/CIFAR-10-images/archive/master.zip", "/tmp")
        os.rename("/tmp/CIFAR-10-images-master", CIFAR_PATH)
        # os.remove("master.zip")


def init_context(context):
    setattr(context.user_data, HDFS_CONNECTION, Config().get_client('dev'))
    setattr(context.user_data, NODE_ID, randint(1, INT_MAX))
    # if not os.path.exists(CIFAR_PATH):
    #     download_images()


def train_epoch(context, event):
    context.logger.info_with('Got invoked',
                             trigger_kind=event.trigger.kind,
                             event_body=event.body)
    body = event.body.decode('utf-8')
    hdfs_client = getattr(context.user_data, HDFS_CONNECTION)
    if body == "train":
        if not os.path.exists(CIFAR_PATH):
            download_images()
        try:
            hdfs_client.download('models/' + AVERAGED_MODEL_NAME, AVERAGED_MODEL_NAME, overwrite=True)
        except HdfsError:
            print("file does not exist")

        history = train(steps_per_epoch=1, batch_size=32)

        save_model_to_hdfs(hdfs_client, getattr(context.user_data, NODE_ID))
        return history.history['loss']
    elif body == "average":
        weights = calculate_average_weights(hdfs_client)
        model = resnet50_model(10)
        model.set_weights(weights)
        model.save(MODEL_WEIGHTS_PATH)
        save_model_to_hdfs(hdfs_client, getattr(context.user_data, NODE_ID), name=AVERAGED_MODEL_NAME)
        return "averaging successful"
    else:
        return "invalid command"


def hdfs_test(client: Client):
    with open('/tmp/model.json') as model, client.write('test/model.json', overwrite=True, encoding='utf-8') as writer:
        writer.write(model.read())


def save_model_to_hdfs(client: Client, node_id: str, name=""):
    if os.path.exists(MODEL_WEIGHTS_PATH):
        if name == "":
            name = 'weights-' + str(node_id) + '.json'
        print("saving " + name)
        with open(MODEL_WEIGHTS_PATH, 'rb') as model, client.write('models/' + name, overwrite=True) as writer:
            writer.write(model.read())


def main():
    body = "average"
    hdfs_client = Config().get_client('dev')
    save_model_to_hdfs(hdfs_client, "1245", name=AVERAGED_MODEL_NAME)


def load_models_from_storage(client: Client):
    files = client.list('models')
    models = list()
    for file in files:
        client.download('models/' + file, file, overwrite=True)
        models.append(keras.models.load_model(file))
    # with client.read('models/' + file) as reader:
    # models.append(keras.models.load_model(MODEL_WEIGHTS_PATH))
    return models


def load_weights_from_storage(client: Client):
    models = load_models_from_storage(client)
    weights = [model.get_weights() for model in models]
    return weights


def calculate_average_weights(client: Client):
    weights = load_weights_from_storage(client)
    new_weights = list()

    # Average all weights
    # Thanks Marcin Mozejko and ursusminimus! https://stackoverflow.com/a/48212579 https://stackoverflow.com/a/59324995
    for weights_list_tuple in zip(*weights):
        new_weights.append(np.array([np.array(w).mean(axis=0) for w in zip(*weights_list_tuple)]))
    return new_weights


# Tell python to run main method
if __name__ == '__main__': main()
