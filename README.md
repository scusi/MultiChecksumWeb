# README

MultiChecksumWeb is a Webapplication variabt of MultiChecksum optimized for docker

## Download

 git clone https://github.com/scusi/MultiChecksumWeb.git

## Building

 cd MultiChecksumWeb.git
 CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .
 docker build -t yourname/multichecksumweb

## Running docker container

 docker run --publish 80:80 yourname/multichecksumweb

Point your browser to http://127.0.0.1/
