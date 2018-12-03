FROM golang:1.10-alpine as builder
RUN apk add --no-cache ca-certificates git && \
      wget -qO/go/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && \
      chmod +x /go/bin/dep

ENV PROJECT github.com/census-ecosystem/opencensus-experiments/interoptest/src/testcoordinator
WORKDIR /go/src/$PROJECT

# restore dependencies
COPY Gopkg.* ./
RUN dep ensure --vendor-only -v

COPY . .
RUN go build -gcflags='-N -l' -o /testcoordinator .

FROM alpine as release
RUN apk add --no-cache ca-certificates

COPY --from=builder /testcoordinator /testcoordinator
EXPOSE 10000
ENTRYPOINT ["/testcoordinator"]
