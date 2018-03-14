package k8

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type ServiceWatcher interface {
	Client() kubernetes.Interface
	WatchAddService(interface{})
	WatchUpdateService(interface{}, interface{})
	WatchDeleteService(interface{})
}

type EndpointsWatcher interface {
	Client() kubernetes.Interface
	WatchAddEndpoints(interface{})
	WatchUpdateEndpoints(interface{}, interface{})
	WatchDeleteEndpoints(interface{})
}

func WatchEndpoints(w EndpointsWatcher, filters func(options *metav1.ListOptions)) {
	restClient := w.Client().CoreV1().RESTClient()

	watchlist := cache.NewFilteredListWatchFromClient(restClient, k8Endpoints, metav1.NamespaceAll, filters)
	_, informer := cache.NewInformer(watchlist, &v1.Endpoints{}, defaultResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    w.WatchAddEndpoints,
			UpdateFunc: w.WatchUpdateEndpoints,
			DeleteFunc: w.WatchDeleteEndpoints,
		},
	)
	stop := make(chan struct{})
	go informer.Run(stop)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logger.Info("Shutting down watcher")
	close(stop)
}

// Need to have a channel just for sending over events for services
func WatchServices(w ServiceWatcher, filters func(options *metav1.ListOptions)) {
	restClient := w.Client().CoreV1().RESTClient()
	watchlist := cache.NewFilteredListWatchFromClient(restClient, k8Services, metav1.NamespaceAll, filters)
	_, informer := cache.NewInformer(watchlist, &v1.Service{}, defaultResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    w.WatchAddService,
			UpdateFunc: w.WatchUpdateService,
			DeleteFunc: w.WatchDeleteService,
		},
	)
	stop := make(chan struct{})
	go informer.Run(stop)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logger.Info("Shutting down watcher")
	close(stop)
}
