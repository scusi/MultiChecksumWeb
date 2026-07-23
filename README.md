# MultiChecksumWeb

MultiChecksumWeb is a web application variant of [MultiChecksum](https://github.com/scusi/MultiChecksum) optimized for use with [Docker](http://docker.com) containers. It calculates MD5, SHA-1, SHA-224, SHA-256, and SHA-512 checksums for uploaded files.

## Features

- **Multiple Hash Algorithms**: MD5, SHA-1, SHA-224, SHA-256, SHA-512
- **Drag & Drop Upload**: Modern, user-friendly interface with drag and drop support
- **Progress Indicator**: Shows upload progress for large files
- **Copy to Clipboard**: One-click copy for each checksum
- **Download All**: Download all checksums as a text file
- **Privacy Focused**: Files are never stored on disk, only processed in memory
- **Security**: File size limits, XSS protection, and proper error handling
- **Minimal Docker Image**: Uses `scratch` base image for maximum security

## Use with Docker

### Pull and Run

```bash
docker pull scusi/multichecksumweb
docker run --publish 80:80 -d scusi/multichecksumweb
```

Point your browser to your _docker_ip_ or http://localhost:80

### Build Your Own Docker Image from Source

#### Download Sources

```bash
git clone https://github.com/scusi/MultiChecksumWeb.git
cd MultiChecksumWeb
```

#### Building a Binary

The binary is already included in the repository, but you can rebuild it with:

```bash
CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -o MultiChecksumWeb .
```

#### Building a Docker Image

```bash
docker build -t yourname/multichecksumweb .
```

#### Running Your Docker Container

```bash
docker run --publish 80:80 -d yourname/multichecksumweb
```

Point your browser to http://127.0.0.1/

## Configuration

The application can be configured using environment variables:

- `PORT`: The port to listen on (default: 8080 in development, 80 in Docker)

Example:

```bash
docker run -e PORT=8080 --publish 8080:8080 -d yourname/multichecksumweb
```

## Security Features

- **File Size Limit**: Maximum upload size is 100MB to prevent DoS attacks
- **XSS Protection**: All user-provided data (filenames) are properly escaped
- **Memory Safety**: Files are processed in a streaming fashion to minimize memory usage
- **No File Storage**: Uploaded files are never written to disk
- **Minimal Attack Surface**: Docker image uses `scratch` base with only the compiled binary

## API Endpoints

- `GET /` - Upload form
- `POST /do/` - Process uploaded file and return checksums
- `GET /health` - Health check endpoint

## License

This project is open source and available under the same license as [MultiChecksum](https://github.com/scusi/MultiChecksum).
