FROM debian:buster-slim

## --- GOLANG ---
# To make the applicaion compatible with the dv5.1, go version needs to be 1.15 or higher
FROM golang:1.15.6-buster AS go-compiled

COPY . /armiarma

WORKDIR /armiarma/src
RUN go build -o ./bin/app

# --- Install python 3.7 or + in the go image ---
WORKDIR /armiarma

RUN apt-get update && apt-get install -y python3 python3-dev python3-pip

RUN pip3 install -r src/analyzer/requirements.txt

# Arguments coming from the docker call: (1)->Network (2)->Project Name (3)->Time Duration 

ENTRYPOINT ["./armiarma.sh"]

EXPOSE 9020