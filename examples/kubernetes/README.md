# Using Prismic Proxy Cache with Kubernetes

This proxy was made to work with Kubernetes from the start, so you should be able to get a service up and running by applying [`prismic.yml`](prismic.yml):

1. Get a copy of [`prismic.yml`](prismic.yml)
2. Customize it as needed, replacing at least the following values:
    * `<PROXY-HOST>`: the actual host for your proxy, i.e. `content.my-site.com`
    * `<PRISMIC-REPO-NAME>`: name of your repository, which is whatever comes before `.cdn.prismic.io`
3. `kubectl apply -f prismic.yml`

This should give you a deployment, service and ingress to expose your newly created proxy.
