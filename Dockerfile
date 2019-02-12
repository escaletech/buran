FROM alpine

WORKDIR /app

COPY ./dist/buran /app

CMD ["/app/buran"]