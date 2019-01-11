FROM golang:1.11-alpine3.7 as builder
RUN apk add --no-cache ca-certificates git && \
    apk add --no-cache bash git openssh gcc musl-dev

WORKDIR /go/src/

ENV GO111MODULE on
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -gcflags='-N -l' -o /testcoordinator cmd/main.go

FROM alpine as release
RUN apk add --no-cache ca-certificates curl

COPY --from=builder /testcoordinator /testcoordinator
EXPOSE 10000
ENTRYPOINT ["/testcoordinator"]

