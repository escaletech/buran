# Buran 🚀

Lightning-fast proxy cache for Prismic Headless API CMS

## Getting started

### Configuration

Use the following environment variables to configure the server:

* `PORT` (default: `3000`) - Port on which the application will listen
* `BACKEND_URL` (*required*, ex: `http://your-repo.cdn.prismic.io`) - URL of your Prismic API backend
* `REDIS_URL` (default: `redis://localhost`) - Redis connection URL, if you choose to use Redis as a cache
* `CACHE_PROVIDER` (default: `redis`, values: `redis`, `memory`) - which cache provider implementation to use

### Running in Kubernetes

This proxy was built with Kubernetes in mind, so check out the [Kubernetes example](/examples/kubernetes/README.md) to see how to deploy it.

### Running with Docker

The [Docker image](https://hub.docker.com/r/escaletech/buran/tags) used in the Kubernetes example is available for you to use in any way you choose:

```sh
$ docker run --name prismic \
    --env BACKEND_URL='http://<your-repo>.cdn.prismic.io' \
    --env CACHE_PROVIDER='memory' \
    -p 3000:3000 escaletech/buran
```


## Development

### Requirements

* Go 1.11
* GNU Make

### Running

After cloning the project, run:

```sh
$ BACKEND_URL='http://<your-repo>.cdn.prismic.io' \
    CACHE_PROVIDER='memory' \
    make
```

And you should see:

```sh
INFO[0000] listening on port 3000
```

Then you can make a request to the local server and see the Prismic API response:

```sh
$ curl localhost:3000/api/v2
```

If a request is served from cache, you should see the header `X-From-Cache: 1`.


## Contributing

Contributions are welcome! So feel free to open an issue or submit a pull request.
