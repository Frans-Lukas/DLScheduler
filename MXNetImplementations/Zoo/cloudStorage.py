from google.cloud import storage


def upload_blob(bucket_name, source_file_name, destination_blob_name):
    """Uploads a file to the bucket."""
    # The ID of your GCS bucket
    # bucket_name = "your-bucket-name"
    # The path to your file to upload
    # source_file_name = "local/path/to/file"
    # The ID of your GCS object
    # destination_blob_name = "storage-object-name"

    storage_client = storage.Client()
    bucket = storage_client.bucket(bucket_name)
    blob = bucket.blob(destination_blob_name)

    blob.upload_from_filename(source_file_name)

    print(
        "File {} uploaded to {}.".format(
            source_file_name, destination_blob_name
        )
    )


def download_blob(bucket_name, source_blob_name, destination_file_name):
    """Downloads a blob from the bucket."""
    # bucket_name = "your-bucket-name"
    # source_blob_name = "storage-object-name"
    # destination_file_name = "local/path/to/file"

    storage_client = storage.Client()

    bucket = storage_client.bucket(bucket_name)
    exists = storage.Blob(bucket=bucket, name=source_blob_name).exists(storage_client)
    if exists:
        # Construct a client side representation of a blob.
        # Note `Bucket.blob` differs from `Bucket.get_blob` as it doesn't retrieve
        # any content from Google Cloud Storage. As we don't need additional data,
        # using `Bucket.blob` is preferred here.
        blob = bucket.blob(source_blob_name)
        blob.download_to_filename(destination_file_name)

        print(
            "Blob {} downloaded to {}.".format(
                source_blob_name, destination_file_name
            )
        )
    else:
        print("file {} does not exist.".format(source_blob_name))


def upload_simple(local_name: str, external_name: str):
    upload_blob("deep-learning-bucket-for-master-project", local_name, external_name)


def download_simple(external_name: str, local_name: str):
    upload_blob("deep-learning-bucket-for-master-project", external_name, local_name)


def main():
    upload_blob("deep-learning-bucket-for-master-project", "cloudStorage.py", "testfile.txt")
    download_blob("deep-learning-bucket-for-master-project", "testfile.txt", "testfile.txt")


if __name__ == '__main__': main()
