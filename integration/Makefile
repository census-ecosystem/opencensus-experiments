test: gen_proto servers clients

gen_proto:
	protoc -I proto/ proto/defs.proto --go_out=plugins=grpc:proto

servers:
	# TODO: Capture these PIDs and kill them
	go run go-grpc-server/main.go &
	go run go-http-server/main.go &

clients:
	go test -v go-grpc-client/*.go
	go test -v go-http-client/*.go
