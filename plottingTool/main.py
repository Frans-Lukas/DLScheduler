import glob
import json
import os

import plotly.express as px
import plotly.graph_objects as go
import psutil
import sys
import kaleido
import statsmodels
import numpy as np

if __name__ == '__main__':
    inputFolderPath = "../JobHandler/output"

    for directory in glob.glob(os.path.join(inputFolderPath, "*")):
        for filename in glob.glob(os.path.join(directory, '*.txt')):
            with open(os.path.join(os.getcwd(), filename), 'r') as f:
                if "baseLine" in filename:
                    dataUnfiltered = f.read()
                    dataFiltered = dataUnfiltered.replace(" | ", ", ")
                    fileDict = json.loads(dataFiltered)

                    y = fileDict.get('lossHistory')

                    if y is not None:
                        fig = px.scatter(x=np.linspace(1, len(y), len(y)), y=y, trendline="lowess")
                        fig.write_image('results/loss' + filename.replace(inputFolderPath, "").replace(".txt", "plot.png"))
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
                        else:
                            color = "LightSeaGreen"

                        sca = px.scatter(x=var.get('epochs'), y=var.get('loss'), trendline="lowess")
                        sca.update_traces(line=dict(color=color))
                        sca.update_traces(marker=dict(color=color))
                        loss.add_trace(sca.data[0])
                        loss.add_trace(sca.data[1])

                        sca = px.scatter(x=var.get('epochs'), y=var.get('time'), trendline="lowess")
                        sca.update_traces(line=dict(color=color))
                        sca.update_traces(marker=dict(color=color))
                        time.add_trace(sca.data[0])
                        time.add_trace(sca.data[1])

                        sca = px.scatter(x=var.get('epochs'), y=var.get('workers'), trendline="lowess")
                        sca.update_traces(line=dict(color=color))
                        sca.update_traces(marker=dict(color=color))
                        workers.add_trace(sca.data[0])
                        workers.add_trace(sca.data[1])

                        sca = px.scatter(x=var.get('epochs'), y=var.get('servers'), trendline="lowess")
                        sca.update_traces(line=dict(color=color))
                        sca.update_traces(marker=dict(color=color))
                        servers.add_trace(sca.data[0])
                        servers.add_trace(sca.data[1])

                        i += 1

                    loss.write_image('results/loss' + filename.replace(inputFolderPath, "").replace(".txt", "plot.png"))
                    time.write_image('results/time' + filename.replace(inputFolderPath, "").replace(".txt", "plot.png"))
                    workers.write_image(
                        'results/workers' + filename.replace(inputFolderPath, "").replace(".txt", "plot.png"))
                    servers.write_image(
                        'results/servers' + filename.replace(inputFolderPath, "").replace(".txt", "plot.png"))
