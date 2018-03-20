package controller

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TODO: These should probably be configurable as flags
	devModeInternalContext = "prototype-general"
	devModeRemoteContext   = "prototype-secure"
)

var (
	logger = logging.Logger()
)

// Coordinate takes two sets of Kubernetes configurations, one for a client that
// interfaces with the cluster that the controller will run on (called internal)
// and one for a client that interfaces with the peered (remote) cluster.
// It then sets up watchers on both the remote and internal cluster.
func Coordinate(internalConf, remoteConf *rest.Config) error {
	ctx := context.Background()
	internalServiceChan, remoteServiceChan := make(chan *k8.ServiceRequest), make(chan *k8.ServiceRequest)
	remoteEndpointsChan := make(chan *k8.EndpointsRequest)

	internalClient, err := kubernetes.NewForConfig(internalConf)
	if err != nil {
		return ferrors.Error(ctx, err)
	}
	internal := k8.NewInternalClient(internalClient, internalServiceChan, remoteServiceChan, remoteEndpointsChan)

	logger.Info("Setting up remote client")
	remoteClient, err := kubernetes.NewForConfig(remoteConf)
	if err != nil {
		return ferrors.Error(ctx, err)
	}

	remote := k8.NewRemoteClient(remoteClient, remoteServiceChan, remoteEndpointsChan)
	stopChan := make(chan struct{})
	coordinateInternal(internal, stopChan)
	coordinateRemote(remote, stopChan)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	logger.Info("Shutting down watchers")
	close(stopChan)
	return nil
}

func coordinateRemote(remote *k8.RemoteClient, stopChan chan struct{}) {
	logger.Info("Watching remote endpoints")
	filter := func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", k8.CrossClusterServiceLabelKey, k8.CrossClusterServiceLabelValue)
	}
	k8.WatchEndpoints(remote, filter, stopChan)
	k8.WatchServices(remote, filter, stopChan)
}

func coordinateInternal(internal *k8.InternalClient, stopChan chan struct{}) {
	filter := func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", k8.CrossClusterServiceLabelKey, k8.CrossClusterServiceLabelValue)
	}
	k8.WatchServices(internal, filter, stopChan)
	go func() {
		internal.HandleRemoteServiceEvents()
	}()
	go func() {
		internal.HandleRemoteEndpointsEvents()
	}()
	go func() {
		internal.HandleInternalServiceEvents()
	}()
}
