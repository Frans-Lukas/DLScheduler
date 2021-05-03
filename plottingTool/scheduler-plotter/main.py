import glob
import json
import os
from pathlib import Path

import pandas as pd

if __name__ == '__main__':
    inputFolderPath = "../../JobHandler/output"

    oneJobDynamic = [
        "single_job_default_scheduler_83_tl.txt",
        "single_job_gang_scheduler_83_tl.txt",
    ]
    oneJobStatic = [
        "single_job_gang_scheduler_83_tl_static_2w_2s.txt",
        "multi_job_default_scheduler_83_tl_static_2w_2s.txt"
    ]
    twoJobsDynamic = [
        "multi_job_default_scheduler_83_tl.txt",
        "multi_job_gang_scheduler_83_tl.txt",
    ]
    twoJobsStatic = [
        "multi_job_default_scheduler_83_tl_static_2w_2s.txt",
        "multi_job_gang_scheduler_83_tl_static_2w_2s.txt",
    ]
    threeJobsDynamic = [
        "multi_job_three_83_tl.txt",
        "multi_job_three_gang_scheduler_83_tl.txt",
    ]
    threeJobsStatic = [
        "multi_job_three_83_tl_static_2w_2s.txt",
        "multi_job_three_gang_scheduler_83_tl_static_2w_2s.txt",
    ]

    dataCouples = [oneJobDynamic, oneJobStatic, twoJobsDynamic, twoJobsStatic, threeJobsDynamic, threeJobsStatic]
    models = ["Cifar10", "LeNet"]
    pd.set_option("display.max_rows", 101)
    pd.set_option("display.max_columns", 101)
    pd.set_option('display.width', 1000)
    for directory in glob.glob(os.path.join(inputFolderPath, "*")):
        Path("results/gang/LeNet").mkdir(parents=True, exist_ok=True)
        Path("results/gang/Cifar10").mkdir(parents=True, exist_ok=True)
        for dataCouple in dataCouples:
            for model in models:
                for fileName in dataCouple:
                    path = os.path.join(inputFolderPath, model, fileName)
                    try:
                        with open(path, 'r') as f:
                            dataUnfiltered = f.read()
                            dataFiltered = dataUnfiltered.replace(" | ", ", ")
                            fileDict = json.loads(dataFiltered)
                            currentModels = [model for i in fileDict]
                            testName = [fileName for i in fileDict]
                            df = pd.DataFrame(fileDict, columns=["totalTime"])
                            df["model"] = currentModels
                            df["test_name"] = testName
                            print(df.head())
                    except FileNotFoundError:
                        print(path + "does not exist")
