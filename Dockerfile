FROM golang:1.7

RUN go get github.com/urfave/cli
RUN go get golang.org/x/crypto/pbkdf2

WORKDIR /go/src/github.com/icebob/icepacker
ADD . /go/src/github.com/icebob/icepacker