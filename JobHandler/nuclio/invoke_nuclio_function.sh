#!/bin/bash

echo '{"ip": "'$2'", "role": "'$3'", "num_worker": '$4', "num_server": '$5'}'

nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'"}'
