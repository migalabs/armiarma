# syntax=docker/dockerfile:1

# chosen buster image for
FROM golang:1.17.3-buster

COPY ./cmd /armiarma/cmd
COPY ./src /armiarma/src
COPY ./go.mod /armiarma
COPY ./go.sum /armiarma
COPY ./main.go /armiarma
COPY ./config.json /armiarma

WORKDIR /armiarma 
RUN go get
RUN go build -o ./armiarma-client
RUN mkdir /armiarma/peerstore

# Crawler exposed Port
EXPOSE 9020 
# Crawler exposed Port for Prometheus data export
EXPOSE 9080
# Arguments coming from the docker call: (1)->armiarma-client (2)->flags
ENTRYPOINT ["/armiarma/armiarma-client"]