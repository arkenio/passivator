FROM       arken/gom-base
MAINTAINER Vladimir PASQUIER <vpasquier@nuxeo.com>

RUN go get github.com/arkenio/passivator
WORKDIR /usr/local/go/src/github.com/arkenio/passivator
RUN gom install
RUN gom test

EXPOSE 7777
ENTRYPOINT ["gogeta", "-etcdAddress", "http://172.17.42.1:4001"]
