## --- GOLANG ---
# To make the applicaion compatible with the dv5.1, go version needs to be 1.15 or higher
FROM golang:1.15.6-buster AS builder

COPY . /armiarma

WORKDIR /armiarma/src
RUN go build -o ./bin/armiarma

# FINAL STAGE -> copy the binary
FROM debian:buster-slim

# --- Install python 3.7 or + in the go image ---
COPY --from=builder /armiarma /armiarma

WORKDIR /armiarma
RUN apt update && apt install -y python3 python3-dev python3-pip
RUN apt install -y curl
RUN apt install -y iputils-ping
RUN pip3 install -r ./src/analyzer/requirements.txt

WORKDIR /armiarma
EXPOSE 9020
# Arguments coming from the docker call: (1)->Network (2)->Project Name (3)->Time Duration 
ENTRYPOINT ["./armiarma.sh"]