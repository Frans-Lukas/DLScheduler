import os

import keras
import numpy as np
from hdfs import Client, HdfsError

AVERAGED_MODEL_NAME = "AVERAGE_MODEL.h5"


class DistributedHelper:
    MODEL_WEIGHTS_PATH = ""

    def __init__(self, client: Client, worker_id: int, max_id: int):
        self.client = client
        self.worker_id = int(worker_id)
        self.max_id = max_id
        self.MODEL_WEIGHTS_PATH = "model-" + str(worker_id) + ".h5"

    def download_models_from_hdfs(self, path: str) -> []:
        """
         Downloads all models from hdfs storage
         Returns: a list of the file names of the downloaded models
         path: Where to store the model files
        """
        files = self.client.list('models')
        # models = list()
        for file in files:
            self.client.download('models/' + file, path + "/" + file, overwrite=True)
            # models.append(keras.models.load_model(path + "/" + file))
        # with self.client.read('models/' + file) as reader:

        return files

    def upload_model_to_hdfs(self, file: str, name="") -> bool:
        if os.path.exists(file):
            if name == "":
                name = 'weights-' + str(self.worker_id) + '.h5'
            print("saving trained model to hdfs: " + name)
            with open(file, 'rb') as model, self.client.write('models/' + name, overwrite=True) as writer:
                writer.write(model.read())
            return True
        return False

    def download_averaged_model(self):
        try:
            self.client.download('models/' + AVERAGED_MODEL_NAME, AVERAGED_MODEL_NAME, overwrite=True)
        except HdfsError:
            print("averaged model does not exist")

    def aggregate_weights(self, model):
        """
        Averages the weights by downloading all models from hdfs storage and calculating the average of each weight.
        :param model: The model the weights are based on
        :return: NaN
        """
        weights = self.__calculate_average_weights()
        model.set_weights(weights)
        # save model locally so that it can be stored in hdfs
        model.save(self.MODEL_WEIGHTS_PATH)
        self.__save_average_model()

    def split_data(self, data: []):
        """
        Splits the given input array into a chunk based on the workers id and the number of workers.
        :param data: The input array to be split into a chunk
        :return: A slice of the data array
        """
        data_size = len(data)
        chunk_size = int(data_size / self.max_id)
        start_data = int(chunk_size * self.worker_id)
        end_data = min(start_data + chunk_size, data_size)
        return data[start_data:end_data]

    def __save_average_model(self):
        self.upload_model_to_hdfs(AVERAGED_MODEL_NAME, name=AVERAGED_MODEL_NAME)

    def __local_model(self):
        return keras.models.load_model(self.MODEL_WEIGHTS_PATH)

    def __calculate_average_weights(self):
        files = self.download_models_from_hdfs("/tmp")
        weights = self.__get_weights_from_models("/tmp", files)
        new_weights = list()

        # Average all weights
        # Thanks Marcin Mozejko and ursusminimus! https://stackoverflow.com/a/48212579 https://stackoverflow.com/a/59324995
        for weights_list_tuple in zip(*weights):
            new_weights.append(np.array([np.array(w).mean(axis=0) for w in zip(*weights_list_tuple)]))
        return new_weights

    def __hdfs_test(client: Client):
        with open('/tmp/model.json') as model, client.write('test/model.json', overwrite=True,
                                                            encoding='utf-8') as writer:
            writer.write(model.read())

    def __get_weights_from_models(self, path, files):
        weights = []
        for file in files:
            weights.append(keras.models.load_model(path + "/" + file).get_weights())
        return weights
