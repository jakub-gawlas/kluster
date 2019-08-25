FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY test/server/server server

CMD ["./server"]
