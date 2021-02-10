#!/bin/bash
cd server
nuctl deploy --path . --registry docker.io/franslukas -n nuclio
cd ../worker
nuctl deploy --path . --registry docker.io/franslukas -n nuclio
nuctl get function -n nuclio