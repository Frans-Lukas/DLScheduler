import glob
import json
import os

import plotly.express as px
import sys

if __name__ == '__main__':
    inputFolderPath = "../JobHandler/output"

    for directory in glob.glob(os.path.join(inputFolderPath, "*")):
        for filename in glob.glob(os.path.join(directory, '*.txt')):
            with open(os.path.join(os.getcwd(), filename), 'r') as f:
                dataUnfiltered = f.read()
                dataFiltered = dataUnfiltered.replace(" | ", ", ")
                fileDict = json.loads(dataFiltered)
                fig = px.line(x=fileDict[0].get('epochs'), y=fileDict[0].get('loss'))
                fig.write_html('results/loss' + filename.replace(inputFolderPath, "").replace(".txt", ".html"),
                               auto_open=False)
                fig = px.line(x=fileDict[0].get('epochs'), y=fileDict[0].get('time'))
                fig.write_html('results/time' + filename.replace(inputFolderPath, "").replace(".txt", ".html"),
                               auto_open=False)
                fig = px.line(x=fileDict[0].get('epochs'), y=fileDict[0].get('workers'))
                fig.write_html('results/workers' + filename.replace(inputFolderPath, "").replace(".txt", ".html"),
                               auto_open=False)
                fig = px.line(x=fileDict[0].get('epochs'), y=fileDict[0].get('servers'))
                fig.write_html('results/servers' + filename.replace(inputFolderPath, "").replace(".txt", ".html"),
                               auto_open=False)
