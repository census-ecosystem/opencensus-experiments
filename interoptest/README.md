# Interoperability Testing

## Goals

OpenCensus is supported in multiple languages. It is normal to have microservices implemented 
in different languages so it's critical that distributed tracing and stats work uniformly and
interoperate with different implementation across languages. The goal of the interop testing
work is to develop a framework to verify the propagation of OpenCensus TraceContexts and 
TagContexts across service written in different languages

## Support Matrix

### Trace Propagation

| Transport | Propagation Format | Go | Java | Python | Nodejs | C++ |
|-----------|--------------------|----|------|--------|--------|-----|
| gRPC | Binary | Y | Y | Y | Y | ? |
| HTTP | B3 | Y | Y | Y | Y | ? |
| HTTP | TraceContext | Y | Y | Y | Y | ? |

### Tag Propagation

|  Transport | Propagation Format | Go | Java | Python | Nodejs | C++ |
|------------|--------------------|----|------|--------|--------|-----|
| gRPC | Binary | Y | Y | Y | Y | ? |
| HTTP | ?? | ? | ? | ? | ? | ? |

## Design
![Interoperability Test Design Diagram][InteroperabilityTestDesignDiagram]

### Controller
Controller is responsible for instantiating all containers and triggering test run by requesting Test Coordinator to run certain tests. 
TODO: Should it monitor Test Coordinator and post the result somewhere?

Implement using skaffold, docker, and shell script/python.

### Test Coordinator

Test Coordinator (TC) manages all the tests. It is the central point of interoperability test. It 
 generates all test requests and creates a report based on the response and trace export from OC 
 Agent.
 
TC will be implemented in Go.
 
#### Test Coordinator Tasks 
- Listens on Port 10000 for all requests (GET/POST).
- Provides following services over gRPC.
  - Service Registration. All Servers registers their service using `register` rpc.
  - Test Service. 
    - Controller can request a test run using 'run' rpc. 
    - Controller can request a result using 'result' rpc.
- Executes test cases in background asynchronously upon receiving run request over gRPC.
- Listens on Port 10001 for trace/stats export from OC Agent.
- Parses trace export from OC Agent and stitches them together to create a trace.
- Validates that there exists a trace for each test case and the trace itself ii valid.
- Creates a report for each test case.
- [TODO] what should be the format of result. Right now it is defined in proto. Should this xml/html/json?

### Server

There is one container (process) for each language. The server acts as a client and server both. Server performs following Tasks.
Listens on Port 10000 + X, where X is designated for each language.

| Server | X | Base Port |
|--------|---|-----------|
| Java | 100 | 10100 |
| Go | 200 | 10200 |
| Nodejs | 300 | 10300 |
| Python | 400 | 10400 |
| C++ | 500 | 10500 |

#### Server Tasks
- Initialize all services that it offers. A Service is a combination of transport and
propagation format. 
  - **Example Services**

    | Server | Service | Port |
    |--------|---------|------|
    | Java | Binary-over-GRPC | 10101 |
    | Java | B3-over-HTTP | 10102 |
    | Java | TraceContext-over-HTTP | 10103 |
  
  - **GRPC Transport**: Use [TestExecutionService][InteroperabilityTestProto] defined in proto.
  - **HTTP Transport**: Use 
     - method: POST 
     - path: `/test/request`
     - Payload: TestRequest protobuf in textformat
- Enable tracing for all services.
- Register all services with Test Coordinator on boot up along with host and port. Port can
be static as long as there are no conflicts.
- Register OC Agent as an exporter for all the services.
- All Service should provide TestService, where it would receive a TestRequest and respond with
 TestResponse. A request may be a single-hop request or a multip-hop request.
 - **Single-hop Request** If a request is single-hop then the server should reply with the response 
 immediately. A request is single-hop if the request does not contain any service-hops (see proto
 for [TestRequest][InteroperabilityTestProto]).
 - **Multi-hop Requests** If a request is multi-hop then the server creates a new request to
 a server as specified in the service-hop (first service-hop if there are more than 1). The server
 Waits for the response and appends it to its own response and sends the reply. See proto for 
 [TestResponse][InteroperabilityTestProto]
  
- Request could contain a list of tags. If they do then tags should be propagated along with trace
 context. [TODO] Do all plugins provide propagation?

##### Example Request and Response
- TC sends a request to Server A with a list of ServiceHop
```
{id, [ServiceHop B1, ServiceHop C1]}.
``` 
- Server A receives the request and originates a child request to next service
```
{id, [ServiceHop C1]}
```
- Server B then receives the request and originates a child request to ServerC:serviceC4 with an 
```
{id, []}
```
- Server C simply returns a response as there are no more servers left. 
```
{id, [CommonResponseStatus C1]}
```
- Server B returns
```
{id, [CommonResponseStatus B1, CommonResponseStatus C1]}
```
- Server A returns 
```
{id, [CommonResponseStatus A1, CommonResponseStatus B1, CommonResponseStatus C1]}
```

## OC Agent
OC Agent simply receives export from each server and forwards that to TC.

## Protos
All proto definitions used for Interoperability test is defined [here][InteroperabilityTestProto]

[InteroperabilityTestDesignDiagram]: /interoptest/drawings/InteroperabilityTestDesignDiagram.png "Interoperability Design"
[InteroperabilityTestProto]: /interoptest/proto/interoperability_test.proto "Interoperability Test Proto"
