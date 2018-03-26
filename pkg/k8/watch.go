package k8

import (
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	defaultResyncPeriod = 30 * time.Second
	k8Endpoints         = "endpoints"
	k8Services          = "services"
)

type Filter func(options *metav1.ListOptions)

type Watcher interface {
	Add(interface{})
	Update(interface{}, interface{})
	Delete(interface{})
}

// WatchEndpoints watches for endpoint update and delete events
func WatchEndpoints(clientset kubernetes.Interface, w Watcher, filters Filter, stopChan <-chan struct{}) {
	restClient := clientset.CoreV1().RESTClient()
	watchlist := cache.NewFilteredListWatchFromClient(restClient, k8Endpoints, metav1.NamespaceAll, filters)
	_, informer := cache.NewInformer(watchlist, &v1.Endpoints{}, defaultResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    w.Add,
			UpdateFunc: w.Update,
			DeleteFunc: w.Delete,
		},
	)
	go informer.Run(stopChan)
}

// WatchServices watches for service add, update, and delete events
func WatchServices(clientset kubernetes.Interface, w Watcher, filters func(options *metav1.ListOptions), stopChan <-chan struct{}) {
	restClient := clientset.CoreV1().RESTClient()
	watchlist := cache.NewFilteredListWatchFromClient(restClient, k8Services, metav1.NamespaceAll, filters)
	_, informer := cache.NewInformer(watchlist, &v1.Service{}, defaultResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    w.Add,
			UpdateFunc: w.Update,
			DeleteFunc: w.Delete,
		},
	)
	go informer.Run(stopChan)
}
