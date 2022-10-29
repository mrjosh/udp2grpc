FROM golang:1.19.1-alpine AS builder

LABEL maintainer="Alireza Josheghani <josheghani.dev@gmail.com>"

RUN apk update && \
  apk add make git

WORKDIR /udp2grpc
ADD . /udp2grpc
RUN make

FROM alpine:3.16.2

COPY --from=builder /udp2grpc/bin/utg /usr/local/bin/utg
EXPOSE 52935

ENTRYPOINT ["/usr/local/bin/utg"]
