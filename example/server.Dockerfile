# syntax = docker/dockerfile:experimental

# build image
FROM golang:1.12-stretch as builder

WORKDIR /app/
COPY  .. .

ENV CGO_ENABLED=0
ENV GOOS=linux
RUN --mount=type=cache,target=/go/pkg/mod go build -a -installsuffix cgo -o server test/server/main.go

# result image
FROM centurylink/ca-certs

WORKDIR /root/

COPY --from=builder /app/server server

CMD ["./server"]
