FROM       arken/gom-base
MAINTAINER Vladimir PASQUIER <vpasquier@nuxeo.com>

RUN go get github.com/arkenio/passivator
WORKDIR /usr/local/go/src/github.com/arkenio/passivator
RUN gom install
RUN gom test

RUN wget --no-verbose https://github.com/coreos/fleet/releases/download/v0.4.0/fleet-v0.4.0-linux-amd64.tar.gz
RUN tar -v -C /tmp -xzf fleet-v0.4.0-linux-amd64.tar.gz
RUN cp /tmp/fleet-v0.4.0-linux-amd64/fleetctl /usr/bin/fleetctl

EXPOSE 7777
ENTRYPOINT ["passivator", "-etcdAddress", "http://172.17.42.1:4001", "-cronDuration", "5", "-passiveLimitDuration", "12","-etcdEndPoint", "http://172.17.42.1:4001", "-alsologtostderr", "true"]
