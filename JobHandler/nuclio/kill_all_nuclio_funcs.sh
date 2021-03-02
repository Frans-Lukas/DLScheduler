#!/bin/bash


echo "make sure to run script as sudo"

#BLACK MAGIC FUCKERY
x=`sudo nuctl get functions -n nuclio | grep -o -P '(?<=nuclio).*(?=default)' | awk '{gsub(/^[ \t \|]+| \|[ \t]+$/,""); print $0}'`
while IFS= read -r line ; do
  sudo nuctl delete function $line -n nuclio;
done <<< "$x"