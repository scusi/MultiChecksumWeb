# README

MultiChecksumWeb is a Webapplication variant of [MultiChecksum](https://github.com/scusi/MultiChecksum) optimized for useing with [docker](http://docker.com) containers

## Download

	git clone https://github.com/scusi/MultiChecksumWeb.git

## Building a binary

 	cd MultiChecksumWeb.git
 	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .

## building a docker image

 	docker build -t yourname/multichecksumweb

## Running docker container

 	docker run --publish 80:80 yourname/multichecksumweb

Point your browser to http://127.0.0.1/
