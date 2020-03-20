FROM golang:1.14.1-stretch

WORKDIR $GOPATH/src/server

COPY . ./

RUN go build

ENTRYPOINT ["./server"]
