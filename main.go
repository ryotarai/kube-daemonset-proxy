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
		Watcher: watcher,
		PodPort: options.PodPort,
		Title:   options.Title,
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
	PodPort          int
}

func parseFlags() (*FlagOptions, error) {
	os.Getenv("")
	opt := &FlagOptions{}
	flag.StringVar(&opt.Namespace, "namespace", "", "Namespace Pods exist in")
	flag.StringVar(&opt.LabelSelectorRaw, "label-selector", "", `Label selector e.g. "a=b,c=d"`)
	flag.StringVar(&opt.ListenAddr, "listen-addr", ":8080", `Address to listen on`)
	flag.StringVar(&opt.Title, "title", "", `Title in index page`)
	flag.IntVar(&opt.PodPort, "pod-port", -1, "Pod port")
	flag.Parse()

	if opt.Namespace == "" {
		return nil, fmt.Errorf("-namespace is not set")
	}
	if opt.PodPort == -1 {
		return nil, fmt.Errorf("-pod-port is not set")
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
