#!/bin/bash

go get -u github.com/hybridgroup/gobot

go get -u contrib.go.opencensus.io/exporter/stackdriver

go get -u go get -u go.opencensus.io

sudo apt-get install sshpass

cd $(go env GOPATH)/src/github.com/census-ecosystem/oepncensus-experiments/go/iot/

chmod u+x ./run.sh
