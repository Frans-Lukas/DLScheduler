import json
import os

import matplotlib.pyplot as plt
import pandas as pd
from matplotlib import gridspec

if __name__ == '__main__':
    inputFolderPath = "../../JobHandler/output2"

    oneJobDynamic = [
        "single_job_default_scheduler_83_tl.txt",
        "single_job_gang_scheduler_83_tl.txt",
    ]
    oneJobStatic = [
        "single_job_default_scheduler_83_tl_static_2w_2s.txt",
        "single_job_gang_scheduler_83_tl_static_2w_2s.txt",
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

    title = [
        "ID=1, M=R, J=1, C=D",
        "ID=2, M=R, J=1, C=S",
        "ID=3, M=R, J=2, C=D",
        "ID=4, M=R, J=2, C=S",
        "ID=5, M=R, J=3, C=D",
        "ID=6, M=R, J=3, C=S",
        "ID=7, M=L, J=1, C=D",
        "ID=8, M=L, J=1, C=S",
        "ID=9, M=L, J=2, C=D",
        "ID=10, M=L, J=2, C=S",
        "ID=11, M=L, J=3, C=D",
        "ID=12 ,M=L, J=3, C=S",
    ]

    dataCouples = [oneJobDynamic, oneJobStatic, twoJobsDynamic, twoJobsStatic, threeJobsDynamic, threeJobsStatic]
    models = ["Cifar10", "LeNet"]
    pd.set_option("display.max_rows", 101)
    pd.set_option("display.max_columns", 101)
    pd.set_option('display.width', 100000)
    pd.set_option("max_colwidth", 400)

    fig = plt.figure(figsize=(10, 8))
    fig.suptitle("Distributed deep learning training times\n with different configurations", fontsize=18)
    fig.text(0.5, 0.04, 'Job Id (+0.5 skew for default scheduler)', ha='center')
    fig.text(0.04, 0.5, 'Job runtime (s)', va='center', rotation='vertical')

    fig2 = plt.figure(figsize=(10, 8))
    fig2.suptitle("Distributed deep learning epoch times\n with different configurations", fontsize=18)
    fig2.text(0.5, 0.04, 'Epoch #', ha='center')
    fig2.text(0.04, 0.5, 'Epoch runtime (s)', va='center', rotation='vertical')

    fig3 = plt.figure(figsize=(10, 8))
    fig3.suptitle("Distributed deep learning epoch times for\nvarious jobs using the gang scheduler", fontsize=18)
    fig3.text(0.5, 0.04, 'Epoch #', ha='center')
    fig3.text(0.04, 0.5, 'Epoch runtime (s)', va='center', rotation='vertical')

    fig4 = plt.figure(figsize=(10, 8))
    fig4.suptitle("Distributed deep learning epoch times for\nvarious jobs using the default scheduler", fontsize=18)
    fig4.text(0.5, 0.04, 'Epoch #', ha='center')
    fig4.text(0.04, 0.5, 'Epoch runtime (s)', va='center', rotation='vertical')
    # plt.legend([""])
    outer = gridspec.GridSpec(3, 4, wspace=0.5, hspace=0.5)
    outer2 = gridspec.GridSpec(3, 4, wspace=0.5, hspace=0.5)
    outer3 = gridspec.GridSpec(3, 4, wspace=0.5, hspace=0.5)
    outer4 = gridspec.GridSpec(3, 4, wspace=0.5, hspace=0.5)

    legendExists = False
    threeFourPrinted = False
    k = 0
    currIdNum = 1
    fileNameToId = {}
    # b1_full = None
    # b2_full = None
    for model in models:
        for dataCouple in dataCouples:
            # print(dataCouple)
            result = pd.DataFrame(columns=["totalTime", "model", "testName", "schedType", "id"])
            result_line = pd.DataFrame(columns=["epochs", "time", "model", "testName", "nrOfEpochs", "schedType", "id"])
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
                        df_line = pd.DataFrame(fileDict, columns=["epochs", "time", "nrOfEpochs"])
                        # print(df_line)
                        df = pd.DataFrame(fileDict, columns=["totalTime"])
                        df["model"] = currentModels
                        df["schedType"] = schedTypes
                        df["testName"] = testName
                        df["id"] = ids
                        result = pd.concat([result, df])
                        result_line = pd.concat([result_line, df_line])

                except FileNotFoundError:
                    print(path + "does not exist")
            inner = gridspec.GridSpecFromSubplotSpec(1, 1, subplot_spec=outer[k])
            inner2 = gridspec.GridSpecFromSubplotSpec(1, 1, subplot_spec=outer2[k])
            inner3 = gridspec.GridSpecFromSubplotSpec(1, 1, subplot_spec=outer3[k])
            inner4 = gridspec.GridSpecFromSubplotSpec(1, 1, subplot_spec=outer4[k])
            ax = plt.Subplot(fig, inner[0])
            ax2 = plt.Subplot(fig2, inner2[0])
            ax3 = plt.Subplot(fig3, inner3[0])
            ax4 = plt.Subplot(fig4, inner4[0])
            # print(result.head())
            ax.title.set_text(title[k])
            ax2.title.set_text(title[k])
            ax3.title.set_text(title[k])
            ax4.title.set_text(title[k])

            b_heights = result[result.schedType == "default"]["totalTime"]
            a_heights = result[result.schedType == "gang"]["totalTime"]
            b_bins = [i for i, _ in enumerate(b_heights)]
            a_bins = [i + 0.5 for i, _ in enumerate(a_heights)]
            print(k)
            print(b_heights.mean())
            print(a_heights.mean())

            bins = result[result.schedType == "gang"]["id"]

            width = max(max(b_bins) if (len(b_bins) > 0) else 0, max(a_bins) if len(a_bins) > 0 else 0) / (
                    len(a_bins) + len(b_bins))

            gangSchedTimes = result_line[result.schedType == "gang"]
            defaultSchedTimes = result_line[result.schedType == "default"]

            gangTimes = []
            gangEpochs = []
            markers = ["x", "o", "*"]
            b12 = None
            b22 = None
            for j, (gangEpoch, gangTime) in enumerate(zip(gangSchedTimes["epochs"], gangSchedTimes["time"])):
                gangEpochs += gangEpoch
                gangTimes += gangTime
                b12 = ax2.scatter(gangEpoch, gangTime, c='cornflowerblue')
                b13 = ax3.scatter(gangEpoch, gangTime, c='cornflowerblue', marker=markers[j])
            defaultTimes = []
            defaultEpochs = []
            for j, (defaultEpoch, defaultTime) in enumerate(
                    zip(defaultSchedTimes["epochs"], defaultSchedTimes["time"])):
                defaultEpochs += defaultEpoch
                defaultTimes += defaultTime
                b22 = ax2.scatter(defaultEpoch, defaultTime, c='seagreen', marker=markers[j])
                b14 = ax4.scatter(defaultEpoch, defaultTime, c='seagreen', marker=markers[j])

            b1 = ax.bar(a_bins, a_heights, width=width, facecolor='cornflowerblue')

            b2 = ax.bar(b_bins, b_heights, width=width, facecolor='seagreen')

            if not legendExists:
                leg = ax.legend([b1, b2], ["gang scheduler", "default scheduler"])
                leg2 = ax2.legend([b12, b22], ["gang scheduler", "default scheduler"])
                legendExists = True
                bb = leg.get_bbox_to_anchor().inverse_transformed(ax.transAxes)
                bb2 = leg2.get_bbox_to_anchor().inverse_transformed(ax2.transAxes)
                xOffset = -0.5
                yOffset = 0.55
                bb.x0 += xOffset
                bb.x1 += xOffset
                bb.y0 += yOffset
                bb.y1 += yOffset
                bb2.x0 += xOffset
                bb2.x1 += xOffset
                bb2.y0 += yOffset
                bb2.y1 += yOffset
                leg.set_bbox_to_anchor(bb, transform=ax.transAxes)
                leg2.set_bbox_to_anchor(bb2, transform=ax2.transAxes)

            k += 1
            fig.add_subplot(ax)
            fig2.add_subplot(ax2)
            fig3.add_subplot(ax3)
            fig4.add_subplot(ax4)
            # if len(b_bins) > 0:
            #     b1_full = b1
            # if len(a_bins) > 0:
            #     b2_full = b2

    # for i in fileNameToId:
    #     print(str(i) + ": " + str(fileNameToId[i]))

    # plt.legend([b1_full, b2_full], ["gang scheduler", "default scheduler"])
    fig.show()
    fig2.show()
    fig3.show()
    fig4.show()
