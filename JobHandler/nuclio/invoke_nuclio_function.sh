#!/bin/bash

echo '{"ip": "'$2'", "role": "'$3'", "num_workers": '$4', "num_servers": '$5'}'


nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_workers": '$4', "num_servers": '$5'}'