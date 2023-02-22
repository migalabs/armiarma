# syntax=docker/dockerfile:1

# chosen buster image for
FROM golang:1.17.13-buster AS builder

COPY ./ /armiarma

WORKDIR /armiarma 
#RUN make dependencies
RUN make build

# FINAL STAGE -> copy the binary and few config files
FROM debian:buster-slim

RUN mkdir /crawler
COPY --from=builder /armiarma/build/ /crawler

# Crawler exposed Port
EXPOSE 9020 
# Crawler exposed Port for Prometheus data export
EXPOSE 9080
# Arguments coming from the docker call: (1)->armiarma-client (2)->flags
ENTRYPOINT ["/crawler/armiarma"]
