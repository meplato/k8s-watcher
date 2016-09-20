# Kubernetes Watcher

This is a very simple tool that watches the endpoints of a
Kubernetes service.

## Pre-requisites

You need to have [Docker](https://www.docker.com/) installed.

## Build and run

To build the binaries locally and run `k8s-watcher` outside of your container, type:

```sh
$ make build
$ ./k8s-watcher -kubeconfig=./my-kubeconfig.json -service=greeter-server -namespace=default
{"buildTag":"","buildTime":"","caller":"main.go:60","kubeconfig":"./kubeconfig.example.json","msg":"Watcher starting","namespace":"default","service":"greeter-server","ts":"2016-09-20T10:48:16+02:00","version":"dev"}
{"caller":"main.go:161","msg":"greeter-server: 172.17.0.2:50051, 172.17.0.3:50051\n","ts":"2016-09-20T10:48:21+02:00"}
{"caller":"main.go:161","msg":"greeter-server: 172.17.0.2:50051, 172.17.0.3:50051\n","ts":"2016-09-20T10:48:26+02:00"}
...
```

You can specify a kubeconfig file to configure connection settings.
The `kubeconfig.example.json` is just an example of a kubeconfig file;
use `kubectl config view` to get a list of all your current cluster
configurations.

Start new pods of the service to see how `k8s-watcher` will find the new endpoints.

## Build and run container

To build the Docker container and run it inside the Kubernetes cluster, type:

```sh
$ VERSION=0.1.0 make container
$ docker images | grep k8s-watcher
meplato/k8s-watcher       0.1.0               3d9ad5bc0927        2 seconds ago       31.4 MB
```

You can now run the image locally by using the `deloyments/k8s-watcher.yaml` file.
Edit that file if you want different versions of `k8s-watcher`.

```sh
$ kubectl create -f deployments/k8s-watcher.yaml
$ kubectl get po --watch
... wait until the pod is ready
$ kubectl logs -f <k8s-watcher-pod>
```

To stop the watcher, run:

```sh
$ kubectl delete -f deployments/k8s-watcher.yaml
```

# License

MIT
