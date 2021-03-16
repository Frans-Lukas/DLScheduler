import os

import numpy as np
from hdfs import Client, HdfsError
from tensorflow.python.keras import Model




class DistributedHelper:
    MODEL_WEIGHTS_PATH = ""
    AVERAGED_MODEL_NAME = ""

    def __init__(self, client: Client, worker_id: int, max_id: int, average_model_name: str, job_id: str):
        self.client = client
        self.worker_id = int(worker_id)
        self.job_id = job_id
        self.max_id = max_id
        self.MODEL_WEIGHTS_PATH = "model-" + str(worker_id) + ".h5"
        self.AVERAGED_MODEL_NAME = average_model_name

    def download_weights_from_hdfs(self, path: str) -> []:
        """
         Downloads all models from hdfs storage
         Returns: a list of the file names of the downloaded models
         path: Where to store the model files
        """
        files = self.client.list('models')
        # models = list()
        for file in files:
            if self.__is_weights_file(file):
                print("downloading: " + file)
                self.client.download('models/' + file, path + "/" + file, overwrite=True)
            # models.append(keras.models.load_model(path + "/" + file))
        # with self.client.read('models/' + file) as reader:

        return files

    def __is_weights_file(self, file):
        return "weights" in file and self.job_id in file

    def upload_weights_to_hdfs(self, file: str, name="") -> bool:
        if os.path.exists(file):
            if name == "":
                name = 'weights-' + str(self.worker_id) + '.h5'
            name = self.job_id + name
            print("saving trained model to hdfs: " + name)
            with open(file, 'rb') as model, self.client.write('models/' + name, overwrite=True) as writer:
                writer.write(model.read())
            return True
        else:
            print("file " + file + " does not exist, curr pwd is: " + os.getcwd())
        return False

    def download_averaged_model(self):
        try:
            self.client.download('models/' + self.AVERAGED_MODEL_NAME, self.AVERAGED_MODEL_NAME, overwrite=True)
        except HdfsError:
            print("averaged model does not exist")

    def aggregate_weights(self, model: Model):
        """
        Averages the weights by downloading all models from hdfs storage and calculating the average of each weight.
        :param model: The model the weights are based on
        :return: NaN
        """
        weights = self.__calculate_average_weights(model)
        if len(weights) == 0:
            print("could not average aggregate weights, returning")
            return -1

        model.set_weights(weights)

        # save model locally so that it can be stored in hdfs
        print("saving average weights to local storage with path: " + self.AVERAGED_MODEL_NAME)
        model.save_weights(self.AVERAGED_MODEL_NAME)
        self.__save_average_model()
        return 0

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

    def tmp_save_avg(self):
        self.__save_average_model()

    def __save_average_model(self):
        print("saving average model to storage with name: ", self.AVERAGED_MODEL_NAME)
        self.upload_weights_to_hdfs(self.AVERAGED_MODEL_NAME, name=self.AVERAGED_MODEL_NAME)

    def __local_model(self, model: Model):
        return model.load_weights(self.MODEL_WEIGHTS_PATH)

    def __calculate_average_weights(self, model: Model):
        print("downloading models from hdfs")
        files = self.download_weights_from_hdfs("/tmp")
        print("retrieving weights from downloaded models")
        weights = self.__get_weights_from_models("/tmp", files, model)
        new_weights = list()
        print("weights size: " + str(len(weights)))

        # Average all weights
        # Thanks Marcin Mozejko and ursusminimus! https://stackoverflow.com/a/48212579 https://stackoverflow.com/a/59324995
        for weights_list_tuple in zip(*weights):
            new_weights.append(np.array([np.array(w).mean(axis=0) for w in zip(*weights_list_tuple)]))
        return new_weights

    def __hdfs_test(self, client: Client):
        with open('/tmp/model.json') as model, client.write('test/model.json', overwrite=True,
                                                            encoding='utf-8') as writer:
            writer.write(model.read())

    def __get_weights_from_models(self, path, files, model: Model):
        weights = []
        for file in files:
            # print("reading file: " + file)
            print("checking if " + self.job_id + " is in file " + file)
            if self.__is_weights_file(file):
                print(file + " is weightsfile")
                filePath = path + "/" + file
                if os.path.exists(filePath):
                    model.load_weights(filePath)
                    weights.append(model.get_weights())
                else:
                    print("file " + filePath + " does not exist")
        return weights
