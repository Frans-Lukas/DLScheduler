import json

from hdfs import Config

from MXNetImplementations.LeNet.TensorFlow.distributed_lib import DistributedHelper

HDFS_CONNECTION = 'hdfs_connection'


def init_context(context):
    setattr(context.user_data, HDFS_CONNECTION, Config().get_client('dev'))
    # if not os.path.exists(CIFAR_PATH):
    #     download_images()


def run(context, event):
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
