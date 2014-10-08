# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM busybox 

# Copy the local package files to the container's workspace.
ADD Md5Webserver /Md5Webserver
ADD tmpl /tmpl
#ADD sh /bin/sh

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
#RUN go install github.com/scusi/Md5Webserver

# Run the outyet command by default when the container starts.
ENTRYPOINT /Md5Webserver

# Set environment variable of port to bind app to
ENV PORT 80

# Document that the service listens on port 80.
EXPOSE 80
