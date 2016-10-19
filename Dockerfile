FROM golang:1.6

ADD . /go/src/github.com/sethgrid/fakettp

RUN go install github.com/sethgrid/fakettp

CMD fakettp
