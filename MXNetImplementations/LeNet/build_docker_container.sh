#!/bin/bash


cd worker
sudo nuctl build lenet --path . --registry docker.io/franslukas
