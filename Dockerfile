FROM golang:1.9.1 as builder
WORKDIR /go/src/github.com/luizbafilho/fusis/
COPY .  .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/fusis

FROM ubuntu:xenial
RUN apt-get update -y && apt-get install -y kmod iptables
WORKDIR /root/
COPY --from=builder /go/src/github.com/luizbafilho/fusis/bin/fusis .
CMD ["./fusis", "balancer"]


