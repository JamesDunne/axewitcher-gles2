#!/bin/sh
while true; do
  $@ &
  PID=$!
  inotifywait -e close_write $1
  kill $PID
done
