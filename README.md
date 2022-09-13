# kube-update-api
Simple API to update an existing deployment on a Kubernetes cluster with a new container image.

Runs on port 8080.

## Warning

This implements no authentication or user checks, it expects to sit behind some kind of authenticating layer.  You can use my https://github.com/richardjkendall/basicauth-rproxy as an example.

## Building

Clone the repo, and then run the following:

```bash
cd cmd/kube-api
go build
```

If you want to build the container image then from the repo root run `docker build -t <tag> .`

## How to run it

This is designed to run inside a Kubernetes cluster, although you can run it outside of a cluster and it can pick up cluster credentials from the usual place `~/.kube/config`

If you are running this outside a Kubernetes cluster then give it the `-external` flag.  E.g. `./kube-api -external`.

There are example manifests for running this in a Kubernetes cluster in the `example-manifests` folder.

## How to use

### Get deployment/container details

`GET /deployments/:namespace/:deployment/:container`

### Update deployment/container image

`POST /deployments/:namespace/:deployment/:container`

expects the body of the request to match:

```json
{
  "Image": "image-url:tag"
}
```



