all: gen_proto

gen_proto:
	protoc -I proto/ proto/defs.proto --go_out=plugins=grpc:proto
