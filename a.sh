#!/bin/zsh
for i in $(seq 1 3); do
  seq 11 61 | xargs -P 10 -I {} ping -c 1 -t 1 192.168.15${i}.{} | grep 'bytes from' | awk '{print $4}' | sed 's/://' | head -1
  if [ $(wc -l | tr -d ' ') -ge 1 ]; then
    break
  fi
done