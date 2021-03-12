#!/bin/bash

echo "make sure to run script as sudo"
echo "attempting to delete all pods with substring: '$1'"

#BLACK MAGIC FUCKERY
x=`nuctl get functions -n nuclio | grep -o -P '(?<=nuclio).*(?=default)' | awk '{gsub(/^[ \t \|]+| \|[ \t]+$/,""); print $0}'`
while IFS= read -r line ; do
  if [[ $line == *$1* ]]; then
    nuctl delete function $line -n nuclio;
  fi
done <<< "$x"