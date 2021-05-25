import glob
import json
import os

import plotly.express as px
from plotly.subplots import make_subplots
import plotly.graph_objects as go
import psutil
import sys
import kaleido
import statsmodels
import numpy as np
from pathlib import Path

if __name__ == '__main__':
    inputFolderPath = "../JobHandler/output"
    outputFolderPath = "results"

    for directory in glob.glob(os.path.join(inputFolderPath, "*")):
        plots = []

        Path(outputFolderPath + "/loss/LeNet").mkdir(parents=True, exist_ok=True)
        Path(outputFolderPath + "/loss/Cifar10").mkdir(parents=True, exist_ok=True)
        Path(outputFolderPath + "/time/LeNet").mkdir(parents=True, exist_ok=True)
        Path(outputFolderPath + "/time/Cifar10").mkdir(parents=True, exist_ok=True)
        Path(outputFolderPath + "/workers/LeNet").mkdir(parents=True, exist_ok=True)
        Path(outputFolderPath + "/workers/Cifar10").mkdir(parents=True, exist_ok=True)
        Path(outputFolderPath + "/servers/LeNet").mkdir(parents=True, exist_ok=True)
        Path(outputFolderPath + "/servers/Cifar10").mkdir(parents=True, exist_ok=True)

        for filename in glob.glob(os.path.join(directory, '*.txt')):
            with open(os.path.join(os.getcwd(), filename), 'r') as f:
                if "baseLine" in filename:
                    dataUnfiltered = f.read()
                    dataFiltered = dataUnfiltered.replace(" | ", ", ")
                    fileDict = json.loads(dataFiltered)

                    y = fileDict.get('lossHistory')

                    if y is not None:
                        fig = px.scatter(x=np.linspace(1, len(y), len(y)), y=y, trendline="lowess")
                        fig.update_layout(title='Loss values per epoch', xaxis_title='Epoch', yaxis_title='Loss')
                        fig.write_image(
                            outputFolderPath + '/loss' + filename.replace(inputFolderPath, "").replace(".txt",
                                                                                                       "plot.png"))
                else:
                    dataUnfiltered = f.read()
                    dataFiltered = dataUnfiltered.replace(" | ", ", ")
                    fileDict = json.loads(dataFiltered)

                    loss = px.scatter()
                    time = px.scatter()
                    workers = px.scatter()
                    servers = px.scatter()

                    i = 0
                    for var in fileDict:
                        color = ""
                        if i == 0:
                            color = "RoyalBlue"
                        elif i == 1:
                            color = "LightSeaGreen"
                        else:
                            color = "firebrick"

                        symbol = ""
                        if i == 0:
                            symbol = "circle"
                        elif i == 1:
                            symbol = "x"
                        else:
                            symbol = "diamond"

                        dash = ""
                        if i == 0:
                            dash = "solid"
                        elif i == 1:
                            dash = "dash"
                        else:
                            dash = "dashdot"

                        name = "Job " + str(i + 1)

                        sca = px.scatter(x=var.get('epochs'), y=var.get('loss'), trendline="lowess")
                        sca.update_traces(line=dict(color=color, dash=dash))
                        sca.update_traces(marker=dict(color=color, symbol=symbol))
                        sca.update_traces(name=name, showlegend=True)
                        loss.add_trace(sca.data[0])
                        loss.add_trace(sca.data[1])

                        sca = px.scatter(x=var.get('epochs'), y=[1 / i for i in var.get('time')], trendline="lowess")
                        sca.update_traces(line=dict(color=color, dash=dash))
                        sca.update_traces(marker=dict(color=color, symbol=symbol))
                        sca.update_traces(name=name, showlegend=True)
                        time.add_trace(sca.data[0])
                        time.add_trace(sca.data[1])

                        sca = px.scatter(x=var.get('epochs'), y=var.get('workers'), trendline="lowess")
                        sca.update_traces(line=dict(color=color, dash=dash))
                        sca.update_traces(marker=dict(color=color, symbol=symbol))
                        sca.update_traces(name=name, showlegend=True)
                        workers.add_trace(sca.data[0])
                        workers.add_trace(sca.data[1])

                        sca = px.scatter(x=var.get('epochs'), y=var.get('servers'), trendline="lowess")
                        sca.update_traces(line=dict(color=color, dash=dash))
                        sca.update_traces(marker=dict(color=color, symbol=symbol))
                        sca.update_traces(name=name, showlegend=True)
                        servers.add_trace(sca.data[0])
                        servers.add_trace(sca.data[1])

                        i += 1

                    loss.update_layout(title='Loss values per epoch', xaxis_title='Epoch', yaxis_title='Loss')
                    time.update_layout(title='Epochs/second each epoch', xaxis_title='Epoch',
                                       yaxis_title='Epochs/second')
                    workers.update_layout(title='Active workers per epoch', xaxis_title='Epoch', yaxis_title='Workers')
                    servers.update_layout(title='Active servers per epoch', xaxis_title='Epoch', yaxis_title='Servers')

                    loss.write_image(
                        outputFolderPath + '/loss' + filename.replace(inputFolderPath, "").replace(".txt", "plot.png"))
                    time.write_image(
                        outputFolderPath + '/time' + filename.replace(inputFolderPath, "").replace(".txt", "plot.png"))
                    workers.write_image(
                        outputFolderPath + '/workers' + filename.replace(inputFolderPath, "").replace(".txt",
                                                                                                      "plot.png"))
                    servers.write_image(
                        outputFolderPath + '/servers' + filename.replace(inputFolderPath, "").replace(".txt",
                                                                                                      "plot.png"))

                    plots.append((filename, loss, time, workers, servers))

        singleLoss = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        multiLoss = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        multiLossThree = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        singleTime = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        multiTime = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        multiTimeThree = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        singleWorkers = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        multiWorkers = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        multiWorkersThree = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        singleServers = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        multiServers = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))
        multiServersThree = make_subplots(rows=3, cols=2, subplot_titles=(
            "Default", "Static 1W 1S", "Static 1W 2S", "Static 2W 1S", "Static 2W 2S"))

        defaultWorkersServersLeNet = make_subplots(rows=3, cols=2, subplot_titles=(
            "Single job workers", "Single job servers", "Two jobs workers", "Two jobs servers", "Three jobs workers",
            "Three jobs servers"))
        defaultWorkersServersResNet = make_subplots(rows=3, cols=2, subplot_titles=(
            "Single job workers", "Single job servers", "Two jobs workers", "Two jobs servers", "Three jobs workers",
            "Three jobs servers"))

        for plot in plots:
            if "three" in plot[0] and "gang" not in plot[0]:
                currLossFig = multiLossThree
                currTimeFig = multiTimeThree
                currWorkersFig = multiWorkersThree
                currServersFig = multiServersThree
                workerServerRow = 3
            elif "single" in plot[0] and "gang" not in plot[0]:
                currLossFig = singleLoss
                currTimeFig = singleTime
                currWorkersFig = singleWorkers
                currServersFig = singleServers
                workerServerRow = 1
            elif "multi" in plot[0] and "gang" not in plot[0]:
                currLossFig = multiLoss
                currTimeFig = multiTime
                currWorkersFig = multiWorkers
                currServersFig = multiServers
                workerServerRow = 2
            else:
                continue

            if "1w_1s" in plot[0]:
                row = 1
                col = 2
            elif "1w_2s" in plot[0]:
                row = 2
                col = 1
            elif "2w_1s" in plot[0]:
                row = 2
                col = 2
            elif "2w_2s" in plot[0]:
                row = 3
                col = 1
            elif "repeated" in plot[0]:
                continue
            else:
                row = 1
                col = 1

            if row != 1 or col != 1:
                plot[1].update_traces(showlegend=False)
                plot[2].update_traces(showlegend=False)
                plot[3].update_traces(showlegend=False)
                plot[4].update_traces(showlegend=False)

            if row == 1 and col == 1:
                plot[3].update_traces(showlegend=False)
                if workerServerRow != 3:
                    plot[4].update_traces(showlegend=False)
                if "Cifar10" in directory:
                    for trace in plot[3].data:
                        defaultWorkersServersResNet.add_trace(trace, row=workerServerRow, col=1)
                    for trace in plot[4].data:
                        defaultWorkersServersResNet.add_trace(trace, row=workerServerRow, col=2)
                if "LeNet" in directory:
                    for trace in plot[3].data:
                        defaultWorkersServersLeNet.add_trace(trace, row=workerServerRow, col=1)
                    for trace in plot[4].data:
                        defaultWorkersServersLeNet.add_trace(trace, row=workerServerRow, col=2)
                plot[3].update_traces(showlegend=True)
                plot[4].update_traces(showlegend=True)

            for trace in plot[1].data:
                currLossFig.add_trace(trace, row=row, col=col)

            for trace in plot[2].data:
                currTimeFig.add_trace(trace, row=row, col=col)

            for trace in plot[3].data:
                currWorkersFig.add_trace(trace, row=row, col=col)

            for trace in plot[4].data:
                currServersFig.add_trace(trace, row=row, col=col)

        if directory.replace(inputFolderPath + "/", "") == "Cifar10":
            model = "ResNet18"
        else:
            model = "LeNet"

        singleTime.update_layout(title_text="Epochs/Second for a single " + model + " job", height=700, width=900)
        singleTime.update_xaxes(title_text="Epoch")
        singleTime.update_yaxes(title_text="Epochs/Second")

        multiTime.update_layout(title_text="Epochs/Second for two " + model + " jobs", height=700, width=900)
        multiTime.update_xaxes(title_text="Epoch")
        multiTime.update_yaxes(title_text="Epochs/Second")

        multiTimeThree.update_layout(title_text="Epochs/Second for three " + model + " jobs", height=700, width=900)
        multiTimeThree.update_xaxes(title_text="Epoch")
        multiTimeThree.update_yaxes(title_text="Epochs/Second")

        singleLoss.update_layout(title_text="Loss for a single " + model + " job", height=700, width=900)
        singleLoss.update_xaxes(title_text="Epoch")
        singleLoss.update_yaxes(title_text="Loss")

        multiLoss.update_layout(title_text="Loss for two " + model + " jobs", height=700, width=900)
        multiLoss.update_xaxes(title_text="Epoch")
        multiLoss.update_yaxes(title_text="Loss")

        multiLossThree.update_layout(title_text="Loss for three " + model + " jobs", height=700, width=900)
        multiLossThree.update_xaxes(title_text="Epoch")
        multiLossThree.update_yaxes(title_text="Loss")

        singleServers.update_layout(title_text="Servers used for a single " + model + " job", height=700, width=900)
        singleServers.update_xaxes(title_text="Epoch")
        singleServers.update_yaxes(title_text="Servers")

        multiServers.update_layout(title_text="Servers used for two " + model + " jobs", height=700, width=900)
        multiServers.update_xaxes(title_text="Epoch")
        multiServers.update_yaxes(title_text="Servers")

        multiServersThree.update_layout(title_text="Servers used for three " + model + " jobs", height=700, width=900)
        multiServersThree.update_xaxes(title_text="Epoch")
        multiServersThree.update_yaxes(title_text="Servers")

        singleWorkers.update_layout(title_text="Workers used for a single " + model + " job", height=700, width=900)
        singleWorkers.update_xaxes(title_text="Epoch")
        singleWorkers.update_yaxes(title_text="Workers")

        multiWorkers.update_layout(title_text="Workers used for two " + model + " jobs", height=700, width=900)
        multiWorkers.update_xaxes(title_text="Epoch")
        multiWorkers.update_yaxes(title_text="Workers")

        multiWorkersThree.update_layout(title_text="Workers used for three " + model + " jobs", height=700, width=900)
        multiWorkersThree.update_xaxes(title_text="Epoch")
        multiWorkersThree.update_yaxes(title_text="Workers")

        defaultWorkersServersLeNet.update_layout(title_text="Workers and servers used for default LeNet jobs",
                                                 height=700, width=900)
        defaultWorkersServersLeNet.update_xaxes(title_text="Epoch")
        defaultWorkersServersLeNet.update_yaxes(title_text="Workers", col=1)
        defaultWorkersServersLeNet.update_yaxes(title_text="Servers", col=2)

        defaultWorkersServersResNet.update_layout(title_text="Workers and servers used for default ResNet18 jobs",
                                                  height=700, width=900)
        defaultWorkersServersResNet.update_xaxes(title_text="Epoch")
        defaultWorkersServersResNet.update_yaxes(title_text="Workers", col=1)
        defaultWorkersServersResNet.update_yaxes(title_text="Servers", col=2)

        singleLoss.write_image(
            outputFolderPath + '/loss' + directory.replace(inputFolderPath, "") + "/singleLossCombined.png")
        multiLoss.write_image(
            outputFolderPath + '/loss' + directory.replace(inputFolderPath, "") + "/multiLossCombined.png")
        multiLossThree.write_image(
            outputFolderPath + '/loss' + directory.replace(inputFolderPath, "") + "/multiThreeLossCombined.png")
        singleTime.write_image(
            outputFolderPath + '/time' + directory.replace(inputFolderPath, "") + "/singleTimeCombined.png")
        multiTime.write_image(
            outputFolderPath + '/time' + directory.replace(inputFolderPath, "") + "/multiTimeCombined.png")
        multiTimeThree.write_image(
            outputFolderPath + '/time' + directory.replace(inputFolderPath, "") + "/multiTimeThreeCombined.png")
        singleWorkers.write_image(
            outputFolderPath + '/workers' + directory.replace(inputFolderPath, "") + "/singleWorkersCombined.png")
        multiWorkers.write_image(
            outputFolderPath + '/workers' + directory.replace(inputFolderPath, "") + "/multiWorkersCombined.png")
        multiWorkersThree.write_image(
            outputFolderPath + '/workers' + directory.replace(inputFolderPath, "") + "/multiWorkersThreeCombined.png")
        singleServers.write_image(
            outputFolderPath + '/servers' + directory.replace(inputFolderPath, "") + "/singleServersCombined.png")
        multiServers.write_image(
            outputFolderPath + '/servers' + directory.replace(inputFolderPath, "") + "/multiServersCombined.png")
        multiServersThree.write_image(
            outputFolderPath + '/servers' + directory.replace(inputFolderPath, "") + "/multiServersThreeCombined.png")

        if len(defaultWorkersServersResNet.data) != 0:
            defaultWorkersServersResNet.write_image(outputFolderPath + "/defaultWorkersServersResNet.png")
        if len(defaultWorkersServersLeNet.data) != 0:
            defaultWorkersServersLeNet.write_image(outputFolderPath + "/defaultWorkersServersLeNet.png")
