# Buran 🚀

Lightning-fast proxy cache for Prismic Headless API CMS

## Table of contents

- [Buran 🚀](#buran-%F0%9F%9A%80)
  - [Table of contents](#table-of-contents)
  - [Why Buran](#why-buran)
    - [The problem](#the-problem)
    - [Benchmark](#benchmark)
    - [Solution](#solution)
  - [Getting started](#getting-started)
    - [Configuration](#configuration)
    - [Running in Kubernetes](#running-in-kubernetes)
    - [Running with Docker](#running-with-docker)
  - [Development](#development)
    - [Requirements](#requirements)
    - [Running](#running)
  - [Contributing](#contributing)

## Why Buran

### The problem

Prismic is an amazing Headless CMS, with features such as **experiments**, **previews** and **planned releases**. To do so, it uses a versioning model apparently very inspired by how Git works.

For instance, retrieving a document usually consists of two steps:

```sh
# 1. Look for the master reference
ref=`curl http://your-repo.cdn.prismic.io/api/v2 | jq -r '.refs[] | select(.isMasterRef == true) | .ref'`

# 2. Query documents based on reference
curl -g "http://your-repo.cdn.prismic.io/api/v2/documents/search?ref=$ref&q=[[at(document.type, \"home_page\")]]"
```

There are two problems here:
1. The first request is never cached by CDN, so requests originating far from Prismic servers are hurt by latency
2. CDNs are fast, but local cache is faster


### Benchmark

Sample measurements from running the script above from different locations (times in milliseconds):

| avg      | min   | max     |                                                                   |
| -------- | ----- | ------- | ----------------------------------------------------------------- |
| **10.8** |   8   |  **30** | Call **proxy** from inside Kubernetes **cluster**                 |
|   20.7   |  16   |    34   | Call **proxy** from GCP instance in **southamerica-east-1**       |
|   46.3   | **6** |   223   | Call **Prismic CDN** from AWS instance in **us-east-1**           |
|   86.2   |  20   |   329   | Call **Prismic CDN** from GCP instance in **southamerica-east-1** |

**Note**: the Redis instance used for cache has low network performance, so its latency could also
be improved.

**Running the benchmark**
1. Run `npm install axios`
2. Edit `configs` in `benchmark.js` to point to the desired endpoints
3. Run `node benchmark.js <target>`, where `<target>` is one of `prismic`, `buran-remote` or `buran-local`


### Solution

Buran solves the latency problem by adding a cache layer (with Redis in the example). The first
call (to get the master reference) has its cache invalidated whenever content is published, exempting the need to invalidate each query.


## Getting started

### Configuration

Use the following environment variables to configure the server:

* `PORT` (default: `3000`) - Port on which the application will listen
* `BACKEND_URL` (*required*, ex: `http://your-repo.cdn.prismic.io`) - URL of your Prismic API backend
* `REDIS_URL` (default: `redis://localhost`) - Redis connection URL, if you choose to use Redis as a cache
* `CACHE_PROVIDER` (default: `memory`, values: `redis`, `memory`) - which cache provider implementation to use

### Running in Kubernetes

This proxy was built with Kubernetes in mind, so check out the [Kubernetes example](/examples/kubernetes/README.md) to see how to deploy it.

### Running with Docker

The [Docker image](https://hub.docker.com/r/escaletech/buran/tags) used in the Kubernetes example is available for you to use in any way you choose:

```sh
$ docker run --name prismic \
    --env BACKEND_URL='http://<your-repo>.cdn.prismic.io' \
    -p 3000:3000 escaletech/buran
```


## Development

### Requirements

* Go 1.13
* GNU Make

### Running

After cloning the project, run:

```sh
$ BACKEND_URL='http://<your-repo>.cdn.prismic.io' make
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
