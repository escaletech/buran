FROM alpine

WORKDIR /app

COPY ./dist/prismic-proxy /app

CMD ["/app/prismic-proxy"]