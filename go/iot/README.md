### Installation

`go get -u github.com/census-ecosystem/opencensus-experiments`

### Configuration 
To run the script, please do the following instructions first 

`cd $(go env GOPATH)/src/github.com/census-ecosystem/opencensus-experiments`

`chmod u+x ./configure.sh` 

`./configure.sh rasoberry-id raspberry-ip-address raspberry-ssh-password` 

### Instructions 
After all the above, you could run the following command 

`./run.sh raspberry-id raspberry-ip-address raspberry-ssh-password` 

Then the generated main binary file would run on the remote raspberry pi. 

Note: The default raspberry id for the raspberry pi is `pi` 
