### Introduction

This sub-project, as a prototype application of openCensus on the IoT industry, builds a demo for applying the
openCensus framework on the IoT platform. In this project, we use the Raspberry Pi to collect data from some sensors and
export them to the back-end stackDriver server for visualization and persistence.

### Installation

`go get -u github.com/census-ecosystem/opencensus-experiments`

### Configuration

To run the script, please do the following instructions first 

`cd $(go env GOPATH)/src/github.com/census-ecosystem/opencensus-experiments/go/iot/sensor`

`chmod u+x ./configure.sh` 

`./configure.sh raspberry-id raspberry-ip-address raspberry-ssh-password`

### Instructions

After all the above, you could run the following command 

`./run.sh raspberry-id raspberry-ip-address raspberry-ssh-password`

Then the generated main binary file would run on the remote raspberry pi. 

Note: The default raspberry id for the raspberry pi is `pi` 
