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

# Build the binary with debug information removed
RUN go build -ldflags '-w -s' -o cli client/main.go

# Final image
FROM gcr.io/distroless/base
COPY ./service.pem /
COPY --from=builder /go/src/github.com/nleiva/grpc-tls/cli /

CMD ["/cli", "-id", "$ID"]
