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
echo "9. Default scheduler single tenant 3 worker 3 server"
echo "10. Default scheduler multi tenant 3 worker 3 server"
echo "11. Default scheduler single tenant 2 worker 1 server"
echo "12. Default scheduler multi tenant 2 worker 1 server"
echo "13. Default scheduler single tenant 3 worker 1 server"
echo "14. Default scheduler multi tenant 3 worker 1 server"
echo "15. Default scheduler single tenant 1 worker 2 server"
echo "16. Default scheduler multi tenant 1 worker 2 server"
echo ""


if [[ $# -ge 2 ]]; then
  choice=$1
  model=$2
else
  echo -n "Enter test choice:"
  read choice

  echo "1. LeNet model used"
  echo "2. Cifar10 model used"
  echo -n "Enter model choice:"
  read model
fi

case $model in
1)
  model="LeNet"
  ;;
2)
  model="Cifar10"
  ;;
*)
  echo "invalid selection"
  ;;
esac

echo $choice
echo $model


echo "killing all nuclio functions"
./nuclio/kill_all_nuclio_funcs.sh

case $choice in
1)
  echo "Starting default scheduler single tenant"
  go run main.go input/$model/singleTenant83.json output/$model/single_job_default_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
2)
  echo "Starting default scheduler multi tenant"
  go run main.go input/$model/twoTenant83.json output/$model/multi_job_default_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
3)
  echo "Starting gang scheduler single tenant"
  go run main.go input/$model/singleTenant83.json output/$model/single_job_gang_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
4)
  echo "Starting gang scheduler multi tenant"
  go run main.go input/$model/twoTenant83.json output/$model/multi_job_gang_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
5)
  echo "Starting default scheduler single tenant static 1w 1s"
  go run main.go input/$model/singleTenant83StaticOne.json output/$model/single_job_default_scheduler_83_tl_static_1w_1s.txt /etc/kubernetes/admin.conf
  ;;
6)
  echo "Starting default scheduler multi tenant static 1w 1s"
  go run main.go input/$model/twoTenant83StaticOne.json output/$model/multi_job_default_scheduler_83_tl_static_1w_1s.txt /etc/kubernetes/admin.conf
  ;;
7)
  echo "Starting default scheduler single tenant static 2w 2s"
  go run main.go input/$model/singleTenant83StaticTwo.json output/$model/single_job_default_scheduler_83_tl_static_2w_2s.txt /etc/kubernetes/admin.conf
  ;;
8)
  echo "Starting default scheduler multi tenant static 2w 2s"
  go run main.go input/$model/twoTenant83StaticTwo.json output/$model/multi_job_default_scheduler_83_tl_static_2w_2s.txt /etc/kubernetes/admin.conf
  ;;
9)
  echo "Starting default scheduler single tenant static 3w 3s"
  go run main.go input/$model/singleTenant83StaticThree.json output/$model/single_job_default_scheduler_83_tl_static_3w_3s.txt /etc/kubernetes/admin.conf
  ;;
10)
  echo "Starting default scheduler multi tenant static 3w 3s"
  go run main.go input/$model/twoTenant83StaticThree.json output/$model/multi_job_default_scheduler_83_tl_static_3w_3s.txt /etc/kubernetes/admin.conf
  ;;
11)
  echo "Starting default scheduler single tenant static 2w 1s"
  go run main.go input/$model/singleTenant83StaticTwoOne.json output/$model/single_job_default_scheduler_83_tl_static_2w_1s.txt /etc/kubernetes/admin.conf
  ;;
12)
  echo "Starting default scheduler multi tenant static 2w 1s"
  go run main.go input/$model/twoTenant83StaticTwoOne.json output/$model/multi_job_default_scheduler_83_tl_static_2w_1s.txt /etc/kubernetes/admin.conf
  ;;
13)
  echo "Starting default scheduler single tenant static 3w 1s"
  go run main.go input/$model/singleTenant83StaticThreeOne.json output/$model/single_job_default_scheduler_83_tl_static_3w_1s.txt /etc/kubernetes/admin.conf
  ;;
14)
  echo "Starting default scheduler multi tenant static 3w 1s"
  go run main.go input/$model/twoTenant83StaticThreeOne.json output/$model/multi_job_default_scheduler_83_tl_static_3w_1s.txt /etc/kubernetes/admin.conf
  ;;
15)
  echo "Starting default scheduler single tenant static 1w 2s"
  go run main.go input/$model/singleTenant83StaticOneTwo.json output/$model/single_job_default_scheduler_83_tl_static_1w_2s.txt /etc/kubernetes/admin.conf
  ;;
16)
  echo "Starting default scheduler multi tenant static 1w 2s"
  go run main.go input/$model/twoTenant83StaticOneTwo.json output/$model/multi_job_default_scheduler_83_tl_static_1w_2s.txt /etc/kubernetes/admin.conf
  ;;
17)
  echo "Starting gang scheduler single tenant static 2w 2s"
  go run main.go input/$model/singleTenant83StaticTwo.json output/$model/single_job_gang_scheduler_83_tl_static_2w_2s.txt /etc/kubernetes/admin.conf
  ;;
18)
  echo "Starting gang scheduler multi tenant 2w 2s"
  go run main.go input/$model/twoTenant83StaticTwo.json output/$model/multi_job_gang_scheduler_83_tl_static_2w_2s.txt /etc/kubernetes/admin.conf
  ;;
20)
  echo "Starting gang scheduler (three) multi tenant static 2w 2s"
  go run main.go input/$model/threeTenant83StaticTwo.json output/$model/multi_job_three_gang_scheduler_83_tl_static_2w_2s.txt /etc/kubernetes/admin.conf
  ;;
21)
  echo "Starting gang scheduler (three) multi tenant"
  go run main.go input/$model/threeTenant83.json output/$model/multi_job_three_gang_scheduler_83_tl.txt /etc/kubernetes/admin.conf
  ;;
22)
  echo "Starting (three) multi tenant static 2w 2s"
  go run main.go input/$model/threeTenant83StaticTwo.json output/$model/multi_job_three_83_tl_static_2w_2s.txt /etc/kubernetes/admin.conf
  ;;
23)
  echo "Starting (three) multi tenant"
  go run main.go input/$model/threeTenant83.json output/$model/multi_job_three_83_tl.txt /etc/kubernetes/admin.conf
  ;;
24)
  echo "Starting gang scheduler (three) multi tenant"
  go run main.go input/$model/threeTenant83.json output/$model/repeated_three_gang_$3.txt /etc/kubernetes/admin.conf
 ;;
25)
  echo "Starting default scheduler (three) multi tenant"
  go run main.go input/$model/threeTenant83.json output/$model/repeated_three_default_$3.txt /etc/kubernetes/admin.conf
 ;;
*)
  echo "invalid selection"
  ;;
esac


#TODO:
# 1. Run full epochs without serverless restarts as baseline comparison
# 2. Run long-ass DL model with some of the tests
# 3. Run static with 3w and 3s
# 4. Run with Cifar10
