#!/bin/bash
#nuctl deploy --path . --registry docker.io/franslukas -n nuclio --http-trigger-service-type=NodePort
#nuctl deploy --path . --registry docker.io/franslukas -n nuclio --http-trigger-service-type=NodePort
nuctl deploy mxworker --run-image docker.io/franslukas/processor-mxnet_sh:latest --file ./function.yaml -n nuclio --http-trigger-service-type=NodePort
nuctl deploy mxserver --run-image docker.io/franslukas/processor-mxnet_sh:latest --file ./function.yaml -n nuclio --http-trigger-service-type=NodePort
nuctl deploy mxscheduler --run-image docker.io/franslukas/processor-mxnet_sh:latest --file ./function.yaml -n nuclio --http-trigger-service-type=NodePort

#nuctl get function -n nuclio
#../invoke_nuclio_function.sh