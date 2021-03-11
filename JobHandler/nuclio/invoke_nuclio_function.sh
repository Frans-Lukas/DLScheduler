#!/bin/bash

echo '{"worker_id": '$2', "max_id": '$3', "job_type": "'$4'"}'

nuctl invoke $1 -n nuclio -b '{"worker_id": '$2', "max_id": '$3', "job_type": "'$4'"}'
