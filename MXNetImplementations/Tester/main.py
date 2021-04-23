import os


def main():
    print(os.environ['DMLC_PS_ROOT_URI'])
    print(os.environ['DMLC_PS_ROOT_PORT'])
    print(os.environ['DMLC_ROLE'])
    print(os.environ['DMLC_NUM_SERVER'])
    print(os.environ['DMLC_NUM_WORKER'])
    print(os.environ['JOB_ID'])
    print(os.environ['GOOGLE_APPLICATION_CREDENTIALS'])
    print(
        "regexpresultstart{\"loss\":0.9" + ", \"accuracy\":0.9"  + ", \"worker_id\":0}regexpresultend")

if __name__ == '__main__': main()
