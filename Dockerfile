FROM golang:1.9

MAINTAINER Eranga Bandara (erangaeb@gmail.com)

# install dependencies
RUN go get gopkg.in/mgo.v2

# env
ENV ZWITCH_MODE DEV
ENV ZWITCH_NAME senzswitch
ENV ZWITCH_PORT 7070
ENV MONGO_HOST dev.localhost
ENV MONGO_PORT 27017

# copy app
ADD . /app
WORKDIR /app

# build
RUN go build -o build/senz src/*.go

# running on 7070
EXPOSE 7070

# .keys volume
VOLUME ["/app/.keys"]

ENTRYPOINT ["/app/docker-entrypoint.sh"]
