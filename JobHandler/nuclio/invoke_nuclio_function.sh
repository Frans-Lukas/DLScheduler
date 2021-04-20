#!/bin/bash

if [[ $# -eq 7 ]]; then
  echo "invoking with 7 arg"
  echo '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'", "job_id": "'$7'"}'
  nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'", "job_id": "'$7'"}'
elif [[ $# -eq 8 ]]; then
  echo "invoking with 8 arg"
  echo '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'", "job_id": "'$7'", "num_parts": "'$8'"}'
  nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'", "job_id": "'$7'", "num_parts": "'$8'"}'
fi

#echo '{"ip": "'$2'", "role": "'$3'", "num_worker": '$4', "num_server": '$5', "script_path": "'$6'", "num_parts": '$7'}'

#nuctl invoke $1 -n nuclio -b '{"ip": "'$2'", "role": "'$3'", "num_worker": "'$4'", "num_server": "'$5'", "script_path": "'$6'"}'
