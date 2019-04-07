package k8s

import (
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Watcher struct {
	Namespace      string
	Clientset      *kubernetes.Clientset
	LabelSelectors map[string]string
	PodInformer    cache.SharedIndexInformer
}

func NewWatcher(clientset *kubernetes.Clientset, namespace string, labelSelectors map[string]string) (*Watcher, error) {
	w := &Watcher{
		Clientset:      clientset,
		Namespace:      namespace,
		LabelSelectors: labelSelectors,
	}
	err := w.StartInformer()
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Watcher) StartInformer() error {
	factory := informers.NewSharedInformerFactoryWithOptions(w.Clientset, time.Minute*5, informers.WithNamespace(w.Namespace))
	podInformer := factory.Core().V1().Pods().Informer()
	stopCh := make(chan struct{})
	go factory.Start(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, podInformer.HasSynced); !ok {
		return errors.New("failed to sync cache")
	}

	w.PodInformer = podInformer
	return nil
}

func (w *Watcher) Pods() ([]*corev1.Pod, error) {
	objects := w.PodInformer.GetStore().List()
	pods := []*corev1.Pod{}

podLoop:
	for _, o := range objects {
		pod, ok := o.(*corev1.Pod)
		if !ok {
			return nil, fmt.Errorf("type assertion failed corev1.Pod: %v", o)
		}
		if pod == nil {
			continue
		}
		for k, v := range w.LabelSelectors {
			if pod.Labels[k] != v {
				continue podLoop
			}
		}
		pods = append(pods, pod)
	}

	return pods, nil
}
