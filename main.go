package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ryotarai/kube-daemonset-proxy/pkg/handler"
	"github.com/ryotarai/kube-daemonset-proxy/pkg/k8s"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func main() {
	options, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := k8s.NewClientset()
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	watcher, err := k8s.NewWatcher(clientset, options.Namespace, options.LabelSelector)
	if err != nil {
		log.Fatalf("failed to watch Kubernetes: %v", err)
	}

	s, err := handler.New(handler.Options{
		Watcher:     watcher,
		PodPortName: options.PodPortName,
		Title:       options.Title,
	})
	if err != nil {
		log.Fatalf("failed to create HTTP handler: %v", err)
	}

	if err := http.ListenAndServe(options.ListenAddr, s); err != nil {
		log.Fatalf("failed to listen HTTP server: %v", err)
	}
}

type FlagOptions struct {
	Namespace        string
	LabelSelectorRaw string
	ListenAddr       string
	Title            string
	LabelSelector    map[string]string
	PodPortName      string
}

func parseFlags() (*FlagOptions, error) {
	defaultListenAddr := os.Getenv("KUBE_DS_PROXY_LISTEN_ADDR")
	if defaultListenAddr == "" {
		defaultListenAddr = ":8080"
	}

	opt := &FlagOptions{}
	flag.StringVar(&opt.Namespace, "namespace", os.Getenv("KUBE_DS_PROXY_NAMESPACE"), "Namespace Pods exist in (KUBE_DS_PROXY_NAMESPACE in env var)")
	flag.StringVar(&opt.LabelSelectorRaw, "label-selector", os.Getenv("KUBE_DS_PROXY_LABEL_SELECTOR"), `Label selector e.g. "a=b,c=d" (KUBE_DS_PROXY_LABEL_SELECTOR in env var)`)
	flag.StringVar(&opt.ListenAddr, "listen-addr", defaultListenAddr, `Address to listen on (KUBE_DS_PROXY_LISTEN_ADDR in env var)`)
	flag.StringVar(&opt.Title, "title", os.Getenv("KUBE_DS_PROXY_TITLE"), `Title in index page (KUBE_DS_PROXY_TITLE in env var)`)
	flag.StringVar(&opt.PodPortName, "pod-port-name", os.Getenv("KUBE_DS_PROXY_POD_PORT_NAME"), "Name of Pod port (KUBE_DS_PROXY_POD_PORT_NAME in env var)")
	flag.Parse()

	if opt.Namespace == "" {
		return nil, fmt.Errorf("-namespace is not set")
	}
	if opt.PodPortName == "" {
		return nil, fmt.Errorf("-pod-port-name is not set")
	}

	s := map[string]string{}
	if len(opt.LabelSelector) > 0 {
		for _, kv := range strings.Split(opt.LabelSelectorRaw, ",") {
			parts := strings.Split(kv, "=")
			s[parts[0]] = parts[1]
		}
	}
	opt.LabelSelector = s

	return opt, nil
}
