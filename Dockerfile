FROM golang:1.3

ADD . /go/src/github.com/osterzel/clamrest

WORKDIR /go/src/github.com/osterzel/clamrest
RUN go get 
RUN go install github.com/osterzel/clamrest

ENTRYPOINT /go/bin/clamrest

EXPOSE 8080
