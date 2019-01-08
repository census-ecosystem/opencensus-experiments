FROM golang:1.11-alpine3.7 as builder
RUN apk add --no-cache ca-certificates git && \
    apk add --no-cache bash git openssh gcc musl-dev

WORKDIR /go/src/

COPY . .
ENV GOLANG_VERSION 1.11.2
ENV GO111MODULE on
RUN go get ./...
RUN go build -gcflags='-N -l' -o /goservice .

FROM alpine as release
RUN apk add --no-cache ca-certificates curl

COPY --from=builder /goservice /goservice
EXPOSE 10201 10202 10203
ENTRYPOINT ["/goservice"]
