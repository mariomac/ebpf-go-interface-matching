# Build the testserver binary
# Docker command must be invoked from the projec root directory
FROM golang:1.20 as builder

ARG TARGETARCH

ENV GOARCH=$TARGETARCH

WORKDIR /src

# Copy the go manifests and source
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY go.mod go.mod
COPY go.sum go.sum

# Build
RUN go build -o testprog ./cmd/testprog/testprog.go

# Create final image from minimal + built binary
#TODO: use minimal image
FROM ubuntu:latest

WORKDIR /
COPY --from=builder /src/testprog .
USER 0:0

CMD [ "/testprog" ]