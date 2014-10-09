# README

MultiChecksumWeb is a Webapplication variant of [MultiChecksum](https://github.com/scusi/MultiChecksum) optimized for useing with [docker](http://docker.com) containers

## Use with docker

If you have and use docker already you can just download the image and run it.

    docker pull scusi/multichecksumweb
    docker run --publish 80:80 -d scusi/multichecksumweb

Point your browser to your _docker_ip_

## build your own docker image from source

### Download Sources

	git clone https://github.com/scusi/MultiChecksumWeb.git

### Building a binary

This step can be ommited since a binary is shipped with the source. 
However if you want to build a binary to use in the docker image do the following:

 	cd MultiChecksumWeb
 	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .

### building a docker image

With the binary (from the above step) and a Dockerfile you can build your own docker image like this:

 	docker build -t yourname/multichecksumweb .

### Running your docker container

 	docker run --publish 80:80 -d yourname/multichecksumweb

Point your browser to http://127.0.0.1/
