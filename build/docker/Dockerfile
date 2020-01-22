FROM alpine:3.9

USER nobody

ADD build/_output/atomix-consul-raft-node /usr/local/bin/atomix-consul-raft-node

ENTRYPOINT ["atomix-consul-raft-node"]
