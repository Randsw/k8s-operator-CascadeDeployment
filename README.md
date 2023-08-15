# Kubernetes Operator for Cascade Job starting

## Cascade Job controlled by Deployment

## Create manifest

`make manifests`

## Create CRD

`make generate`

## Image build

`docker build -t ghcr.io/randsw/cascadeautooperator .`

## Image push

`docker login`

`docker push ghcr.io/randsw/cascadeautooperator`

## Image deploy

`make deploy IMG=ghcr.io/randsw/cascadeautooperator`

## Delete operator

`make undeploy`

## Deploy using helm

`make install-helm helm-namespace=<your-namespace>`
