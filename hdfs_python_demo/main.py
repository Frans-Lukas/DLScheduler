from hdfs import Client
from hdfs import Config


def save_model(client: Client):
    with open('model.json') as model, client.write('model.json', overwrite=True, encoding='utf-8') as writer:
        from json import dump
        dump(model.read(), writer)


def read_model(client: Client) -> str:
    with client.read('model2.json', encoding='utf-8') as reader:
        from json import load
        model = load(reader)
    return model


# see https://hdfscli.readthedocs.io/en/latest/quickstart.html#configuration for how to configure a client.
if __name__ == '__main__':
    client = Config().get_client('dev')
    save_model(client)
    print(read_model(client))
