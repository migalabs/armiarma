FROM ubuntu:20.04
# we need to download the virtualenv to perform the after-crawling analysis
#RUN apt-get update && apt-get install -y virtualenv
#RUN apt install -y virtualenv

# COPY THE SOURCE CODE TO THE DOCKER 

# --- PYTHON
# Get python3.8 for the analyzer, stable version that was working with the alpha tests
#FROM python:3.8 AS python-build

#COPY ./src/analyzer /armiarma/src/analyzer

WORKDIR /armiarma/src/analyzer
#RUN pip install virtualenv
#RUN python -m virtualenv venv
#ENV VIRTUAL_ENV /env
#ENV PATH /env/bin:$PATH
#RUN pip install -r ./requirements.txt

## --- GOLANG ---
# To make the applicaion compatible with the dv5.1, go version needs to be 1.15 or higher
FROM golang:1.15.6-buster AS go-compiled

COPY . /armiarma

WORKDIR /armiarma/src
RUN go version
RUN go build -o ./bin/app

# --- Install python 3.7 or + in the go image ---
WORKDIR /armiarma

RUN apt-get update && apt-get install -y build-essential python3 python3-dev python3-pip

RUN pip3 install -r src/analyzer/requirements.txt

WORKDIR /armiarma

# Arguments coming from the docker call: (1)->Network (2)->Project Name (3)->Time Duration 

ENTRYPOINT ["./armiarma.sh"]

EXPOSE 9020