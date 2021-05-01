#!/bin/bash

nuctl build mxnet_sh --path . --registry docker.io/franslukas
#nuctl deploy helloworld -n nuclio -p https://raw.githubusercontent.com/nuclio/nuclio/master/hack/examples/golang/helloworld/helloworld.go --http-trigger-service-type=NodePort --registry docker.io/franslukas

# mxnet_train_until_lv