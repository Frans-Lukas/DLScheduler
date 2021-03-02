#!/bin/bash
cd TensorFlow
sudo nuctl deploy --path . --registry docker.io/franslukas -n nuclio
sudo nuctl get function -n nuclio
#../invoke_nuclio_function.sh
