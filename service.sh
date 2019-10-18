#!/bin/bash
killall check
cd /data/check/
nohup /data/check/check  &
#nohup /data/check/check 2>&1 &
