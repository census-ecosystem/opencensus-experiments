#!/bin/bash
if [ "$#" -ne 3 ]; then
  echo "Illegal number of parameters"
  echo "First is the hostname, second is the ipaddress, third is the ssh password"
  exit -1
fi
CGO_ENABLED=1 GOARM=7 GOARCH=arm GOOS=linux CC=arm-linux-gnueabihf-gcc go build ./examples/pi/main.go
sshpass -p $3 ssh $1@$2 "rm -rf /home/pi/main"
sshpass -p $3 scp ./main $1@$2:~/
#TODO Currently PROJECTID should be set by the user himself/herself
sshpass -p $3 ssh $1@$2 "/home/pi/main &"
rm -rf ./main
