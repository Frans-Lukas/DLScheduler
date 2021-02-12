#!/bin/bash
cd own_parameter_server
sudo nuctl deploy --path . --registry docker.io/franslukas -n nuclio
sudo nuctl get function -n nuclio
#../invoke_worker.sh
