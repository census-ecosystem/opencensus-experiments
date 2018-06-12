#!/bin/bash
cd $GOPATH

go get -u github.com/hybridgroup/gobot

go get -u contrib.go.opencensus.io/exporter/stackdriver

go get -u go get -u go.opencensus.io

sudo apt-get install sshpass

cd $GOPATH:/github.com/src/census-ecosystem/oepncensus-experiments/go/iot/

chmod u+x ./run.sh
