import json
import os

import matplotlib.pyplot as plt
import pandas as pd
from matplotlib import gridspec

if __name__ == '__main__':
    inputFolderPath = "../../JobHandler/output"

    oneJobDynamic = [
        "single_job_default_scheduler_83_tl.txt",
        "single_job_gang_scheduler_83_tl.txt",
    ]
    oneJobStatic = [
        "single_job_gang_scheduler_83_tl_static_2w_2s.txt",
        "single_job_default_scheduler_83_tl_static_2w_2s.txt"
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
    pd.set_option('display.width', 100000)
    pd.set_option("max_colwidth", 400)

    fig = plt.figure(figsize=(10,8))
    outer = gridspec.GridSpec(3, 4, wspace=0.5, hspace=0.5)

    k = 0
    currIdNum = 1
    fileNameToId = {}
    for model in models:
        for dataCouple in dataCouples:

            result = pd.DataFrame(columns=["totalTime", "model", "testName", "schedType", "id"])
            for i, fileName in enumerate(dataCouple):
                if fileName not in fileNameToId:
                    fileNameToId[fileName] = currIdNum
                    currIdNum += 1
                path = os.path.join(inputFolderPath, model, fileName)
                try:
                    with open(path, 'r') as f:
                        dataUnfiltered = f.read()
                        dataFiltered = dataUnfiltered.replace(" | ", ", ")
                        fileDict = json.loads(dataFiltered)
                        currentModels = [model for _ in fileDict]
                        testName = [fileName for _ in fileDict]
                        schedType = "gang" if "gang" in fileName else "default"
                        schedTypes = [schedType for _ in fileDict]
                        ids = [i + j for j, _ in enumerate(fileDict)]
                        print(ids)
                        df = pd.DataFrame(fileDict, columns=["totalTime"])
                        df["model"] = currentModels
                        df["schedType"] = schedTypes
                        df["testName"] = testName
                        df["id"] = ids
                        result = pd.concat([result, df])
                except FileNotFoundError:
                    print(path + "does not exist")
            inner = gridspec.GridSpecFromSubplotSpec(1, 1, subplot_spec=outer[k])
            ax = plt.Subplot(fig, inner[0])
            print(result.head())
            ax.title.set_text(str(fileNameToId[dataCouple[0]]) + " vs. " + str(fileNameToId[dataCouple[1]]))
            ax.set_xlabel("Job id")
            ax.set_ylabel("Job runtime (s)")
            b_heights = result[result.schedType == "default"]["totalTime"]
            a_heights = result[result.schedType == "gang"]["totalTime"]
            b_bins = [i + 0.5 for i, _ in enumerate(b_heights)]
            a_bins = [i for i, _ in enumerate(a_heights)]

            bins = result[result.schedType == "gang"]["id"]

            width = 0.4
            print("a_bins")
            print(a_bins)
            print(a_heights)
            b1 = ax.bar(a_bins, a_heights, width=width, facecolor='cornflowerblue')
            print("b_bins")
            print(b_bins)
            print(b_heights)
            b2 = ax.bar(b_bins, b_heights, width=width, facecolor='seagreen')
            # ax.legend([b1, b2], ["gang scheduler", "default scheduler"])

            print()
            print()
            k += 1
            fig.add_subplot(ax)
            handles, labels = ax.get_legend_handles_labels()
            fig.legend(handles, labels, loc='upper center')
    fig.show()