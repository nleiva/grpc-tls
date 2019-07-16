FROM golang:1.12 as builder

# Variables passed at build-time (include defaults)
ARG host=localhost
ARG port=50051

# Needs full path to work with Container's $GOPATH
# However we don't need $GOPATH with Go modules
WORKDIR /go/src/github.com/nleiva/grpc-tls/

# Cache dependencies (if no changes)
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./. .

ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

# Disable CGO and statically link C bindings.
ENV CGO_ENABLED=0

# Compile for Linux
ENV GOOS=linux

# Build the binary with debug information removed.
RUN go build -ldflags '-w -s' -o ser server/main.go

# Final image
FROM alpine

ARG host
ARG port

ENV PORT=$port
ENV HOST=$host

# Only required to run with self-signed certificates.
# COPY ./service.pem /
# COPY ./service.key /

COPY --from=builder /go/src/github.com/nleiva/grpc-tls/ser /

# Install CA certificates, we need them for tls-alpn-01 challenge.
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates

EXPOSE ${PORT}

CMD ["/ser", "-self=false", "-public=true"]
