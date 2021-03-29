#!/bin/bash

if [[ $# -eq 6 ]]; then
  json="{"ip": "'$2'", "role": "'$3'", "num_worker": '$4', "num_server": '$5', "script_path": "'$6'"}"
  echo $1 $json
  nuctl invoke $1 -n nuclio -b "'$json'"
elif [[ $# -eq 7 ]]; then
  json="{"ip": "'$2'", "role": "'$3'", "num_worker": '$4', "num_server": '$5', "script_path": "'$6'", "num_parts": '$7'}"
  nuctl invoke $1 -n nuclio -b "'$json'"
fi

#echo '{"ip": "'$2'", "role": "'$3'", "num_worker": '$4', "num_server": '$5', "script_path": "'$6'", "num_parts": '$7'}'

#nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'"}'
