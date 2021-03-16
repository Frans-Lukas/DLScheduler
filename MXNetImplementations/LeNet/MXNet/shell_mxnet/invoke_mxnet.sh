#!/bin/bash


nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_workers": '$4', "num_servers": '$5'}'