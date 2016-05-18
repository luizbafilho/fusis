FROM debian:jessie

RUN apt-get update
RUN apt-get install -y libnl-3-dev libnl-genl-3-dev ipvsadm
ADD bin/fusis /
ADD fusis.json /

ENTRYPOINT ["/fusis"]


