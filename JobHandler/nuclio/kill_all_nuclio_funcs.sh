#!/bin/bash

#BLACK MAGIC FUCKERY
x=$(nuctl get functions -n nuclio | grep -o -P '(?<=nuclio).*(?=default)' | awk '{gsub(/^[ \t \|]+| \|[ \t]+$/,""); print $0}')
while IFS= read -r line; do
  nuctl delete function $line -n nuclio
done <<<"$x"
y=$(kubectl get pods -n nuclio | grep -o -P '^\S*')
while IFS= read -r line; do
  if [[ $line == "NAME" || $line == *"controller"* || $line == *"dashboard"* ]]; then
    echo "not deleting " $line
  else
    echo "deleting " $line
    kubectl delete pod $line -n nuclio
  fi
done <<<"$y"
