#!/bin/bash

if [[ $# -eq 6 ]]; then
  echo "invoking with 6 arg"
  echo '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'"}'
  echo 'arg 7: '$7''
  nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'"}'
elif [[ $# -eq 7 ]]; then
  echo "invoking with 7 arg"
  echo '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'", "num_parts": "'$7'"}'
  nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'", "num_parts": "'$7'"}'
fi

#echo '{"ip": "'$2'", "role": "'$3'", "num_worker": '$4', "num_server": '$5', "script_path": "'$6'", "num_parts": '$7'}'

#nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'"}'
