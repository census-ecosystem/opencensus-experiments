###Instructions### 
To run the script, please do the following instructions first
go get -u github.com/hybridgroup/gobot
go get -u contrib.go.opencensus.io/exporter/stackdriver
go get -u go get -u go.opencensus.io
sudo apt-get install sshpass
chmod u+x ./run.sh

After all the above, you could run the following command
./run.sh raspberry-id raspberry-ip-address raspberry-ssh-password
Then the generated main binary file would run on the remote raspberry pi.

Note: The default raspberry id for the raspberry pi is `pi`
