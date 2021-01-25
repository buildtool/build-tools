# Quickstart

## Pre requisites:
In order to work with these tools you need:
* Docker - [installation](https://www.docker.com/get-started)
* Kubernetes - if you're using for example [Docker for Mac](https://docs.docker.com/docker-for-mac/install/)
Kubernetes can easily be enabled.

In this example we will build, push and deploy a sample Go project.

Create a Git repository and add a single main package with a http server:

```shell
mkdir quickstart
cd quickstart
git init
```

```go
// main.go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}
```

Add a Dockerfile describing how to build your code:

```dockerfile
# Dockerfile
FROM golang:1.15 as build
WORKDIR /build
ADD . /build

RUN GOOS=linux GOARCH=amd64 go build main.go

FROM debian:buster-slim
COPY --from=build /build/main /
CMD ["/main"]
```

Add the files to source control and commit:
```sh
git add .
git commit -m "Init"
```
Build the docker image:

```sh
build
```

After the build completes you should see output like

```sh
...
Successfully tagged noregistry/quickstart:latest
```

Try to run your newly built docker image:
```sh
docker run --rm -p 8080:8080 noregistry/peter:master
```
and try to access it:
```sh
curl localhost:8080/buildtools
```
You should see a response like:
```sh
Hello, buildtools!
```

Let's try to deploy it to our local Kubernetes cluster, in order for this to work we need a
Kubernetes descriptor file.
Create a `k8s` folder and a file `deploy.yaml`:

```sh
# k8s/deploy.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quickstart
  labels:
    app: quickstart
spec:
  replicas: 1
  selector:
    matchLabels:
      app: quickstart
  template:
    metadata:
      labels:
        app: quickstart
    spec:
      containers:
      - name: quickstart
        imagePullPolicy: IfNotPresent
        image: noregistry/quickstart:${COMMIT}

```

This will create [Kubernetes deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/),
basically starting your application inside the Kubernetes cluster.

```sh
deploy --context docker-desktop
```

```sh
kubectl --context docker-desktop get pods

quickstart-b4c5bc467-lqk6r      1/1     Running        0          3s
```
