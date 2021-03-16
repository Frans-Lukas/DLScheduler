#!/bin/bash
cd MXNet
#nuctl deploy --path . --registry docker.io/franslukas -n nuclio --http-trigger-service-type=NodePort
nuctl deploy mxnet1 --run-image docker.io/franslukas/processor-mxnet:latest --file ./function.yaml -n nuclio --http-trigger-service-type=NodePort
nuctl deploy mxnet2 --run-image docker.io/franslukas/processor-mxnet:latest --file ./function.yaml -n nuclio --http-trigger-service-type=NodePort
nuctl deploy mxnet3 --run-image docker.io/franslukas/processor-mxnet:latest --file ./function.yaml -n nuclio --http-trigger-service-type=NodePort

nuctl get function -n nuclio
#../invoke_nuclio_function.sh