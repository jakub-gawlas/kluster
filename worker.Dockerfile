FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY test/worker/worker worker

CMD ["./worker"]