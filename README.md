# kube-daemonset-proxy

[![Docker Repository on Quay](https://quay.io/repository/ryotarai/kube-daemonset-proxy/status "Docker Repository on Quay")](https://quay.io/repository/ryotarai/kube-daemonset-proxy)

HTTP reverse proxy to Daemonset Pods

![](https://github.com/ryotarai/kube-daemonset-proxy/raw/master/doc/images/index-page.png)

## Endpoints

- `/` shows an index page as above. It includes links to Pods in each nodes.
- `/n/:nodename/...` proxies requests to a Pod in `nodename`.

## Example

This example deploys kube-daemonset-proxy and Netdata on Kind cluster.

```
$ kind create cluster --config example/kind/cluster.yaml --name kube-daemonset-proxy
$ export KUBECONFIG="$(kind get kubeconfig-path --name kube-daemonset-proxy)"
$ docker build . -t kube-daemonset-proxy && kind load docker-image kube-daemonset-proxy --name kube-daemonset-proxy
$ kubectl apply -f example/manifests -R
$ kubectl port-forward -n kube-daemonset-proxy service/kube-daemonset-proxy :80
Forwarding from 127.0.0.1:xxxxx -> 8080
```

Then, visit `http://127.0.0.1:xxxxx` in a browser.
