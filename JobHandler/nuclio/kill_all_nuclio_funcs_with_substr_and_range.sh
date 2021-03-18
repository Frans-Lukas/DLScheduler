#!/bin/bash

echo "make sure to run script as sudo"
echo "attempting to delete all pods with substring: '$1'"

#BLACK MAGIC FUCKERY
x=$(nuctl get functions -n nuclio | grep -o -P '(?<=nuclio).*(?=default)' | awk '{gsub(/^[ \t \|]+| \|[ \t]+$/,""); print $0}')
while IFS= read -r line; do
  if [[ $line == *$1* ]]; then
    echo $line
    # split on job word TODO: what if a function is randomly given the name job? nvm, since we take last element it should work..
    functionId=${line#"$1"}

    #    ($(echo "$line" | tr $1 ' '))
    #    echo "$line"
    echo $functionId

    # FunctionId is an integer && is between inarg 2 and 3.
    if [[ $functionId -ge $2 ]]; then
      echo "deleting $line"
      #      nuctl delete function $line -n nuclio
    fi
  fi
done <<<"$x"
