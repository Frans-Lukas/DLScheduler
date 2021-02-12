#!/bin/bash
cd worker
nuctl deploy --path . --registry docker.io/franslukas -n nuclio
nuctl get function -n nuclio
#../invoke_worker.sh
