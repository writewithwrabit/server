FROM golang:1.13.8-stretch

WORKDIR $GOPATH/src/server

COPY . ./

RUN go build

ENTRYPOINT ["./server"]
