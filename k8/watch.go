package k8

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func WatchServices(k *RemoteClient) error {
	restClient := k.K8Client.CoreV1().RESTClient()
	// There's another list watch that allows us to filter by the labels
	watchlist := cache.NewListWatchFromClient(restClient, k8Services, metav1.NamespaceAll, fields.Everything())

	uninitializedWatchList := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.IncludeUninitialized = true
			options.LabelSelector = crossClusterServiceLabel
			return watchlist.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			options.LabelSelector = crossClusterServiceLabel
			return watchlist.Watch(options)
		},
	}

	_, informer := cache.NewInformer(uninitializedWatchList, &v1.Service{}, defaultResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    k.AddService,
			UpdateFunc: k.UpdateService,
			DeleteFunc: k.DeleteService,
		},
	)
	stop := make(chan struct{})
	go informer.Run(stop)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logger.Info("Shutting down watcher")
	close(stop)
	return nil
}
