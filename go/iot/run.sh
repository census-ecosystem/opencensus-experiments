#!/bin/bash
if [ "$#" -ne 3 ]; then
  echo "Illegal number of parameters"
  exit -1
fi
GOARM=7 GOARCH=arm GOOS=linux go build ./main.go
sshpass -p $3 scp ./main $1@$2:~/
sshpass -p $3 ssh $1@$2 "/home/pi/main"
rm -rf ./main
