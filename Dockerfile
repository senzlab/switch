FROM golang:1.8

MAINTAINER Eranga Bandara (erangaeb@gmail.com)

# install dependencies
RUN go get gopkg.in/mgo.v2

# copy app
ADD . /app
WORKDIR /app

# build
RUN go build -o build/senz src/*.go

# running on 7070
EXPOSE 7070

ENTRYPOINT ["/app/docker-entrypoint.sh"]
