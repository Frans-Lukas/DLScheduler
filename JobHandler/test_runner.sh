#!/bin/bash

echo "Test choices:"
echo "1. Default scheduler single tenant"
echo "2. Default scheduler multi tenant"
echo "3. Gang scheduler single tenant"
echo "4. Gang scheduler multi tenant"
echo "5. Default scheduler single tenant 1 worker 1 server"
echo "6. Default scheduler multi tenant 1 worker 1 server"
echo "7. Default scheduler single tenant 2 worker 2 server"
echo "8. Default scheduler multi tenant 2 worker 2 server"
echo ""
echo -n "Enter test choice:"
read choice

echo "killing all nuclio functions"
./nuclio/kill_all_nuclio_funcs.sh

case $choice in
1)
  echo "Starting default scheduler single tenant"
  go run main.go input/singleTenant83.json output/single_job_default_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
2)
  echo "Starting default scheduler multi tenant"
  go run main.go input/twoTenant83.json output/multi_job_default_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
3)
  echo "Starting gang scheduler single tenant"
  go run main.go input/singleTenant83.json output/single_job_gang_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
4)
  echo "Starting gang scheduler multi tenant"
  go run main.go input/twoTenant83.json output/multi_job_gang_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
5)
  echo "Starting default scheduler single tenant static 1w 1s"
  go run main.go singleTenant83StaticOne.json single_job_default_scheduler_83_tl_static_1w_1s.txt /etc/kubernetes/admin.conf
  ;;
6)
  echo "Starting default scheduler multi tenant static 1w 1s"
  go run main.go twoTenant83StaticOne.json multi_job_default_scheduler_83_tl_static_1w_1s.txt /etc/kubernetes/admin.conf
  ;;
7)
  echo "Starting default scheduler single tenant static 2w 2s"
  go run main.go singleTenant83StaticTwo.json single_job_default_scheduler_83_tl_static_2w_2s.txt /etc/kubernetes/admin.conf
  ;;
8)
  echo "Starting default scheduler multi tenant static 2w 2s"
  go run main.go twoTenant83StaticTwo.json multi_job_default_scheduler_83_tl_static_2w_2s.txt /etc/kubernetes/admin.conf
  ;;
*)
  echo "invalid selection"
  ;;
esac
