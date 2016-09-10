FROM golang:alpine

ENV PORT 3309
EXPOSE 3309

ADD . /go/src/github.com/Xe/hideyhole-site

WORKDIR /go/src/github.com/Xe/hideyhole-site

RUN go build && go install

CMD /go/bin/hideyhole-site
