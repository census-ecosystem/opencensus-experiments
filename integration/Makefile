.PHONY: start_go_servers start_java_servers run_go_client_tests run_java_client_tests
test: gen_proto gradlew_install_opencensus gradlew_install start_go_servers start_java_servers run_go_client_tests run_java_client_tests

gen_proto:
	protoc -I proto/ proto/defs.proto --go_out=plugins=grpc:proto

gradlew_install_opencensus:
	# Install the latest OpenCensus-Java version locally.
	git clone https://github.com/census-instrumentation/opencensus-java.git && cd opencensus-java && ./gradlew install && cd .. && rm -rf opencensus-java

gradlew_install:
	# To use a different OpenCensus and gRPC version, use -PopencensusVersion= and -PgrpcVersion=.
	cd java && ./gradlew installDist && cd ..

start_go_servers:
	# TODO: Capture these PIDs and kill them
	go run go/grpc/server/main.go &
	go run go/http/server/main.go &

start_java_servers:
	# TODO: Capture these PIDs and kill them
	java/build/install/java-interop/bin/GrpcInteropTestServer &
	java/build/install/java-interop/bin/HttpInteropTestServer &

run_go_client_tests:
	go test -v go/grpc/client/*.go
	go test -v go/http/client/*.go

run_java_client_tests:
	java/build/install/java-interop/bin/GrpcInteropTestClient
	java/build/install/java-interop/bin/HttpInteropTestClient

