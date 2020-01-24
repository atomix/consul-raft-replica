FROM alpine:3.9

USER nobody

ADD build/_output/consul-raft-replica /usr/local/bin/consul-raft-replica

ENTRYPOINT ["consul-raft-replica"]
