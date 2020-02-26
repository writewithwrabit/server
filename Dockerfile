FROM golang:1.14.0-stretch

WORKDIR $GOPATH/src/server

COPY . ./

RUN go build

ENTRYPOINT ["./server"]
