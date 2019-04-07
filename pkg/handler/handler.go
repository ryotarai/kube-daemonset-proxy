package handler

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"sort"

	"github.com/ryotarai/kube-daemonset-proxy/pkg/k8s"

	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	_ "github.com/ryotarai/kube-daemonset-proxy/statik"

	corev1 "k8s.io/api/core/v1"
)

func New(options Options) (*Handler, error) {
	s := &Handler{Options: options}
	err := s.prepare()
	if err != nil {
		return nil, err
	}

	return s, nil
}

type Options struct {
	Watcher     *k8s.Watcher
	PodPortName string
	Title       string
}

type Handler struct {
	Options
	router *gin.Engine
}

func (s *Handler) prepare() error {
	statikFS, err := fs.New()
	if err != nil {
		return err
	}

	tmpl := template.New("")
	tmpl, err = s.loadTemplate(tmpl, statikFS, "/templates/index.html.tmpl")
	if err != nil {
		return err
	}

	router := gin.Default()
	router.SetHTMLTemplate(tmpl)
	router.GET("/", s.handleIndex)
	for _, m := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		router.Handle(m, "/n/:nodename/*path", s.handleNodeProxy)
	}
	router.GET("/public/*path", gin.WrapH(http.FileServer(statikFS)))
	s.router = router

	return nil
}

func (s *Handler) loadTemplate(t *template.Template, filesystem http.FileSystem, name string) (*template.Template, error) {
	f, err := filesystem.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	t, err = t.New(name).Parse(string(h))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Handler) handleIndex(c *gin.Context) {
	pods, err := s.Watcher.Pods()
	if err != nil {
		handleError(c, err)
		return
	}

	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Spec.NodeName < pods[j].Spec.NodeName
	})

	c.HTML(200, "/templates/index.html.tmpl", map[string]interface{}{
		"Pods":  pods,
		"Title": s.Title,
	})
}

func (s *Handler) findPortInPod(pod *corev1.Pod) int32 {
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.Name == s.PodPortName {
				return p.ContainerPort
			}
		}
	}
	return -1
}

func handleError(c *gin.Context, err error) {
	log.Printf("error: %v", err)
	c.String(500, "Internal error: %v\n", err)
}

func (s *Handler) handleNodeProxy(c *gin.Context) {
	nodeName := c.Param("nodename")
	path := c.Param("path")

	pods, err := s.Watcher.Pods()
	if err != nil {
		handleError(c, err)
		return
	}

	var pod *corev1.Pod
	for _, p := range pods {
		if p.Spec.NodeName == nodeName {
			pod = p
		}
	}
	if pod == nil {
		c.Status(404)
		return
	}

	port := s.findPortInPod(pod)
	if port < 0 {
		handleError(c, err)
		return
	}

	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = fmt.Sprintf("%s:%d", pod.Status.PodIP, port)
		req.URL.Path = path
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	// TODO: reuse reverseProxy
	p := &httputil.ReverseProxy{Director: director}
	p.ServeHTTP(c.Writer, c.Request)
}
