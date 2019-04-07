
FROM golang:1.12 AS builder
WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .
RUN GO111MODULE=on go mod download

COPY Makefile .
RUN make install-tools

COPY . .
RUN make build && cp bin/kube-daemonset-proxy /usr/bin/kube-daemonset-proxy

###############################################

FROM ubuntu:18.04
COPY --from=builder /usr/bin/kube-daemonset-proxy /usr/bin/kube-daemonset-proxy

ENTRYPOINT ["/usr/bin/kube-daemonset-proxy"]
