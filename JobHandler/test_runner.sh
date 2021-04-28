#!/bin/bash

echo "Test choices:"
echo "1. Default scheduler single tenant"
echo "2. Default scheduler multi tenant"
echo "3. Gang scheduler single tenant"
echo "4. Default scheduler single tenant 1 worker 1 server"
echo "5. Default scheduler multi tenant 1 worker 1 server"
echo "6. Default scheduler single tenant 2 worker 2 server"
echo "7. Default scheduler multi tenant 2 worker 2 server"
echo ""
echo "Enter test choice:"
read choice


#echo "killing all nuclio functions"
#./nuclio/kill_all_nuclio_funcs.sh


case $choice in
  1)
    echo "Starting defaul sched single tenant"

    ;;
  2)
    ;;
  3)
    ;;
  4)
    ;;
  5)
    ;;
  6)
    ;;
  7)
    ;;
  *)
    echo "invalid selection"
    ;;
esac

