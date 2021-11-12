# syntax=docker/dockerfile:1

# chosen buster image for
FROM golang:1.17.3-buster AS builder

COPY ./ /armiarma

WORKDIR /armiarma 
RUN go get
RUN go build -o ./armiarma-client

# FINAL STAGE -> copy the binary and few config files
FROM debian:buster-slim

COPY --from=builder /armiarma/src /armiarma/src
COPY --from=builder /armiarma/config.json /armiarma/config.json
COPY --from=builder /armiarma/armiarma-client /armiarma/armiarma-client

# Generate the peerstore folder where the peerstore and the metrics will be stored
RUN mkdir /armiarma/peerstore


RUN ls -l /armiarma
WORKDIR /armiarma
# Crawler exposed Port
EXPOSE 9020 
# Crawler exposed Port for Prometheus data export
EXPOSE 9080
# Arguments coming from the docker call: (1)->armiarma-client (2)->flags
ENTRYPOINT ["/armiarma-client"]