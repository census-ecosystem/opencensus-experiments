### Introduction

When designing the IoT system structure, developers would always prefer the master/slave topology with the consideration
of the costs. In this topology, there would be multiple slave nodes that consist of cheaper MCUs and connect to the
master node, which would always be more powerful and expensive. In this way, we could improve the hardware resource
utilization and decrease costs since slave nodes supporting simple hardware interface would be adequate for the sub-tasks.

In this project, I design a protocol for the IoT systems with the above topology. It allows the slave nodes to transmit
the collected information to the master node which runs the openCensus framework. Systems that supports this protocol
would have great flexibility since any MCUs could be the slave node if it follows the protocol.

### Installation

`go get -u github.com/census-ecosystem/opencensus-experiments`

### Configuration

To run the script, please do the following instructions first 

`cd $(go env GOPATH)/src/github.com/census-ecosystem/opencensus-experiments/go/iot/protocol`

`chmod u+x ./configure.sh` 

`./configure.sh raspberry-id raspberry-ip-address raspberry-ssh-password`

### Instructions

After all the above, you could run the following command 

`./run.sh raspberry-id raspberry-ip-address raspberry-ssh-password`

Then the generated main binary file would run on the remote raspberry pi. 

Note: The default raspberry id for the raspberry pi is `pi` 
