FROM golang:1.8

MAINTAINER Eranga Bandara (erangaeb@gmail.com)

# copy app
ADD . /app
WORKDIR /app

# running on 7070
EXPOSE 7070

# build
RUN go build ./z.go

# run
CMD ["./z"]
