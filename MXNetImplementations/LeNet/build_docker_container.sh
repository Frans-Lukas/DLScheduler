#!/bin/bash

cd MXNet
nuctl build mxnet --path . --registry docker.io/franslukas
