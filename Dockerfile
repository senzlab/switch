FROM golang:1.9

MAINTAINER Eranga Bandara (erangaeb@gmail.com)

# install dependencies
RUN go get gopkg.in/mgo.v2

# env
ENV SWITCH_MODE DEV
ENV SWITCH_NAME senzswitch
ENV SWITCH_PORT 7070
ENV MONGO_HOST dev.localhost
ENV MONGO_PORT 27017

# copy app
ADD . /app
WORKDIR /app

# build
RUN go build -o build/senz src/*.go

# running on 7070
EXPOSE 7070

ENTRYPOINT ["/app/docker-entrypoint.sh"]
