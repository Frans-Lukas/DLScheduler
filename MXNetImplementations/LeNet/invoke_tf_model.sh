#!/bin/bash
sudo nuctl invoke lenet -n nuclio -b '{"worker_id": 1, "max_id": 5, "job_type": "average"}'
