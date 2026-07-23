# Build stage: Compile the Go application statically
FROM golang:1.21-alpine AS builder

# Set necessary environment variables for static compilation
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Create and change to the app directory
WORKDIR /app

# Copy source files
COPY md5websrv.go .
COPY go.mod .
COPY tmpl ./tmpl

# Build the application statically
RUN go build -ldflags '-w -s' -o /MultiChecksumWeb .

# Final stage: Use scratch for minimal image
FROM scratch

# Copy the statically compiled binary from the builder stage
COPY --from=builder /MultiChecksumWeb /MultiChecksumWeb

# Copy template files
COPY --from=builder /app/tmpl /tmpl

# Set environment variable for port
ENV PORT=80

# Set environment variable for template directory
ENV TEMPLATE_DIR=/tmpl/

# Expose port 80
EXPOSE 80

# Set the entrypoint to run the application
ENTRYPOINT ["/MultiChecksumWeb"]

# Optional: Add labels for maintainer information
LABEL maintainer="Florian Walther <flw@posteo.de>"
LABEL description="MultiChecksumWeb - A web service for calculating file checksums"
LABEL version="1.0"
