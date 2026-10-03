FROM golang:1.24-alpine AS air
RUN go install github.com/air-verse/air@v1.62.0

FROM phpswoole/swoole:6.2.1-php8.5-alpine
COPY --from=air /go/bin/air /usr/local/bin/air
WORKDIR /var/www
ENTRYPOINT ["air"]
CMD ["-c", ".air.toml"]
