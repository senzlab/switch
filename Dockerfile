FROM golang:1.9

MAINTAINER Eranga Bandara (erangaeb@gmail.com)

# install dependencies
RUN go get gopkg.in/mgo.v2

# env
ENV ZWITCH_MODE DEV
ENV ZWITCH_NAME senzswitch
ENV ZWITCH_PORT 7171
ENV MONGO_HOST dev.localhost
ENV MONGO_PORT 27017
ENV PROMIZE_API https://chainz.com:8443/promizes
ENV USER_API https://chainz.com:8443/users

# copy app
ADD . /app
WORKDIR /app

# build
RUN go build -o build/senz src/*.go

# running on 7171
EXPOSE 7171

# .keys volume
VOLUME ["/app/.keys"]

# .certs volume
VOLUME ["/app/.certs"]

ENTRYPOINT ["/app/docker-entrypoint.sh"]
