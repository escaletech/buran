FROM alpine

RUN apk update && apk add ca-certificates tzdata && rm -rf /var/cache/apk/*

WORKDIR /app

COPY ./dist/buran /app

CMD ["/app/buran"]