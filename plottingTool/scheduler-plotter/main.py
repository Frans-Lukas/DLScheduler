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

    fig = plt.figure(figsize=(10, 8))
    fig.suptitle("Deep learning training time comparisons\n with different configurations", fontsize=18)
    fig.text(0.5, 0.04, 'Job Id (+0.5 skew for default scheduler)', ha='center')
    fig.text(0.04, 0.5, 'Job runtime (s)', va='center', rotation='vertical')
    # plt.legend([""])
    outer = gridspec.GridSpec(3, 4, wspace=0.5, hspace=0.5)

    legendExists = False
    threeFourPrinted = False
    k = 0
    currIdNum = 1
    fileNameToId = {}
    # b1_full = None
    # b2_full = None
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
            # print(result.head())
            ax.title.set_text(str(fileNameToId[dataCouple[0]]) + " vs. " + str(fileNameToId[dataCouple[1]]))
            # ax.set_xlabel("Job id")
            # ax.set_ylabel("Job runtime (s)")
            b_heights = result[result.schedType == "default"]["totalTime"]
            a_heights = result[result.schedType == "gang"]["totalTime"]
            b_bins = [i + 0.5 for i, _ in enumerate(b_heights)]
            a_bins = [i for i, _ in enumerate(a_heights)]

            bins = result[result.schedType == "gang"]["id"]

            width = max(max(b_bins) if (len(b_bins) > 0) else 0, max(a_bins) if len(a_bins) > 0 else 0) / (len(a_bins) + len(b_bins))
            # if fileNameToId[dataCouple[0]] == 3 and fileNameToId[dataCouple[1]] == 4:
            #     width = 0.001

            b1 = ax.bar(a_bins, a_heights, width=width, facecolor='cornflowerblue')

            b2 = ax.bar(b_bins, b_heights, width=width, facecolor='seagreen')
            if fileNameToId[dataCouple[0]] == 3 and fileNameToId[dataCouple[1]] == 4:
                print("a_bins")
                print(a_bins)
                print(a_heights)
                print("b_bins")
                print(b_bins)
                print(b_heights)
                print()
                print()
                if not threeFourPrinted:
                    threeFourPrinted = True
                    continue

            if not legendExists:
                leg = ax.legend([b1, b2], ["gang scheduler", "default scheduler"])
                legendExists = True
                bb = leg.get_bbox_to_anchor().inverse_transformed(ax.transAxes)
                xOffset = -0.5
                yOffset = 0.55
                bb.x0 += xOffset
                bb.x1 += xOffset
                bb.y0 += yOffset
                bb.y1 += yOffset
                leg.set_bbox_to_anchor(bb, transform=ax.transAxes)

            k += 1
            fig.add_subplot(ax)
            # if len(b_bins) > 0:
            #     b1_full = b1
            # if len(a_bins) > 0:
            #     b2_full = b2

    for i in fileNameToId:
        print(str(i) + ": " + str(fileNameToId[i]))

    # plt.legend([b1_full, b2_full], ["gang scheduler", "default scheduler"])
    fig.show()
