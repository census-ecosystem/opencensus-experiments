test: gen_proto gradlew_install go_servers java_servers go_clients java_client

gen_proto:
	protoc -I src/main/proto/ src/main/proto/defs.proto --go_out=plugins=grpc:proto

gradlew_install:
	./gradlew installDist

go_servers:
	# TODO: Capture these PIDs and kill them
	go run go-grpc-server/main.go &
	go run go-http-server/main.go &

java_servers:
	# TODO: Capture these PIDs and kill them
	build/install/java-interop/bin/GrpcInteropTestServer &

go_clients:
	go test -v go-grpc-client/*.go
	go test -v go-http-client/*.go

java_client:
	build/install/java-interop/bin/GrpcInteropTestClient 9800
	build/install/java-interop/bin/GrpcInteropTestClient 9801
