# Interop testing

## Scope

Goal: OpenCensus can propagate traces and tags between processes written in different programming languages with complete fidelity using multiple transports.

* [DONE] P0: Initially only Java and Go(Go implementation completed with https://github.com/census-instrumentation/opencensus-experiments/pull/3) 
* [PARTIAL] P0: Covers gRPC and HTTP transports
* [DONE] P0: trace context propagation
* [PARTIAL] P01: Tag propagation (gRPC done, HTTP unimplemented)
* [NOT STARTED] P1: Stackdriver trace integration testing
* [NOT STARTED] P2: C++, Python

### Out of scope

* Any stats
* Integration testing with Zipkin, Prometheus, Jaeger
* Performance testing

## Design

The service definition will be as follows:

```proto
service EchoService {
        rpc echo(EchoRequest) returns (EchoResponse) {}
}

message EchoRequest {
}

message EchoResponse {
        bytes trace_id = 1;
        bytes span_id = 2;
        int32 trace_options = 3;
        bytes tags_blob = 4;
}
```

> TODO: these are slightly out of date

For HTTP, the request will be an empty body POST and the response will be a single protobuf EchoResponse message. The EchoResponse will be sent back by the server as a JSON response that’s deserializable as the protobuf message. Please note that tags_blob is the byteslice from encoding 

The client behavior will be:

1. Set up expected root span
1. Add expected tags
1. Invoke EchoService.echo / HTTP POST
1. Examine the EchoResponse for the expected values, fail the test if the values in the response don't match the origin values we sent the request
1. The following propagation formats will be tested for HTTP:
    * Google / Stackdriver
    * B3

## Dependency management

> TODO: For Go we will use dep

Java will pin to the latest published version at the time and use Gradle for building

## Running

```
$ make test
```

Environment variables

Variable|Default|Details
---|---|---
OPENCENSUS_GO_GRPC_INTEGRATION_TEST_SERVER_ADDR|:9800|The address on which gRPC clients will find the Go gRPC server
OPENCENSUS_JAVA_GRPC_INTEGRATION_TEST_SERVER_ADDR|:9801|The address on which gRPC clients will find the Java gRPC server
OPENCENSUS_GO_HTTP_INTEGRATION_TEST_SERVER_ADDR|:9900|The address on which HTTP clients will find the Go HTTP server
OPENCENSUS_JAVA_HTTP_INTEGRATION_TEST_SERVER_ADDR|:9901|The address on which HTTP clients will find the Java HTTP server
