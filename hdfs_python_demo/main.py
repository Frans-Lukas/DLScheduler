import os

from hdfs import Client, InsecureClient

# overwrites model with name "model.json" to hdfs storage
# model is taken from local model.json file located in same folder as this script.
ADDRESS = "http://192.168.10.205:9870"


def save_model(client: Client):
    with open('/tmp/model.json') as model, client.write('model.json', overwrite=True, encoding='utf-8') as writer:
        writer.write(model.read())


def test_hdfs(context, event):
    context.logger.info_with('Got invoked',
                             trigger_kind=event.trigger.kind,
                             event_body=event.body,
                             some_env=os.environ.get('MY_ENV_VALUE'))
    # onlyfiles = [f for f in os.listdir(".") if os.path.isfile(os.path.join(".", f))]
    # return onlyfiles
    # with open("/tmp/model.json", "r") as f:
    #     return f.read()
    try:
        client.write('model.json', overwrite=True, encoding='utf-8')
        test_client = InsecureClient(ADDRESS, "franslukas")
        save_model(test_client)
    except Exception as e:
        return e
    return read_model(test_client)
    # import getpass
    # return getpass.getuser()


# reads model with name "model.json" from hdfs storage
def read_model(client: Client) -> str:
    with client.read('model.json', encoding='utf-8') as reader:
        from json import load
        model = load(reader)
    return model


# host.minikube.internal
# see https://hdfscli.readthedocs.io/en/latest/quickstart.html#configuration for how to configure a client.
# defaults to ~/.hdfscli.cfg
if __name__ == '__main__':
    # client = Config().get_client('dev')
    client = InsecureClient(ADDRESS, "franslukas")
    save_model(client)
    print(read_model(client))
