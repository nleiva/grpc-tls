FROM golang:1.12 as builder

# Needs full path to work with Container's $GOPATH
# However we don't need $GOPATH with Go modules
WORKDIR /go/src/github.com/nleiva/grpc-tls/
COPY ./. .

# https://github.com/chemidy/smallest-secured-golang-docker-image/blob/master/Dockerfile
# RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org
ENV PORT ${PORT}
ENV HOST ${HOST}

# Disable CGO and statically link C bindings.
ENV CGO_ENABLED=0

# Compile for Linux
ENV GOOS=linux

# Build the binary with debug information removed
RUN go build -ldflags '-w -s' -o ser server/main.go

# Could be any of these base images I guess
# FROM scratch
# FROM alpine
FROM gcr.io/distroless/base
COPY ./service.pem /
COPY ./service.key /
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/nleiva/grpc-tls/ser /

EXPOSE ${PORT}

CMD ["/ser"]