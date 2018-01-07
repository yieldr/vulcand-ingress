# Vulcand Ingress Controller

A kubernetes [ingress controller](https://kubernetes.io/docs/concepts/services-networking/ingress/) for Mailgun's [vulcand](https://github.com/vulcand/vulcand).

Set up the ingress controller alongside `vulcand`. An example can be found in [examples/ingress-controller.yaml](/yieldr/vulcand-ingress/blob/master/examples/ingress-controller.yaml).

This example uses the public [yieldr/vulcand-ingress](https://hub.docker.com/r/yieldr/vulcand-ingress/) docker image.

### Usage

Start the Ingress Controller using the following command.

```
vulcand-ingress [flags]
```

### Options

```
  -h, --help                  help for vulcand-ingress
      --kubeconfig string     Absolute path to the kubeconfig file. If empty an in-cluster configuration is assumed.
      --namespace string      Namespace in which to watch for resources.
      --selector string       Selector with which to match resources.
      --vulcand-addr string   Vulcand API address. (default "http://localhost:8182")
```
