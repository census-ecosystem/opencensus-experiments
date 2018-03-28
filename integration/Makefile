.PHONY: start_go_servers start_java_servers run_go_client_tests run_java_client_tests
test: gen_proto gradlew_install start_go_servers start_java_servers run_go_client_tests run_java_client_tests

gen_proto:
	protoc -I proto/ proto/defs.proto --go_out=plugins=grpc:proto

gradlew_install:
	# To use a different OpenCensus and gRPC version, use -PopencensusVersion= and -PgrpcVersion=.
	./gradlew installDist

start_go_servers:
	# TODO: Capture these PIDs and kill them
	go run go/grpc/server/main.go &
	go run go/http/server/main.go &

start_java_servers:
	# TODO: Capture these PIDs and kill them
	build/install/java-interop/bin/GrpcInteropTestServer &

run_go_client_tests:
	go test -v go/grpc/client/*.go
	go test -v go/http/client/*.go

run_java_client_tests:
	build/install/java-interop/bin/GrpcInteropTestClient
