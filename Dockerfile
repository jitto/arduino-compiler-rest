FROM golang:1.16.9-stretch as build
COPY go.mod /go/src/test/go.mod
WORKDIR /go/src/test
RUN go mod download

FROM build as install
COPY main.go /go/src/test/main.go
RUN go get test
RUN go install

FROM frolvlad/alpine-glibc
#RUN apk add ca-certificates
COPY --from=install /go/bin/test /usr/bin/arduino-cli-http
WORKDIR /source
RUN arduino-cli-http core install arduino:avr
#ENV USER root
CMD ["arduino-cli-http"]
