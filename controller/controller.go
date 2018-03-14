package controller

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/wearefair/k8-cross-cluster-controller/k8"
	"github.com/wearefair/k8-cross-cluster-controller/utils"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	devModeInternalContext = "prototype-general"
	devModeRemoteContext   = "prototype-secure"
)

var (
	logger = logging.Logger()
)

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

	coordinateInternal(internal)
	coordinateRemote(remote)
	return nil
}

func coordinateRemote(remote *k8.RemoteClient) {
	logger.Info("Watching remote endpoints")
	filter := func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", k8.CrossClusterServiceLabelKey, k8.CrossClusterServiceLabelValue)
	}
	k8.WatchEndpoints(remote, filter)
	k8.WatchServices(remote, filter)
}

func coordinateInternal(internal *k8.InternalClient) {
	filter := func(options *metav1.ListOptions) {}
	k8.WatchServices(internal, filter)
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

func SetupInternalConfig() (*rest.Config, error) {
	var conf *rest.Config
	var err error
	if utils.DevMode() {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{
			CurrentContext: devModeInternalContext,
		}
		localKubeConf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		conf, err = localKubeConf.ClientConfig()
		// TODO: Remove this when done testing locally
		if err != nil {
			return nil, ferrors.Error(context.Background(), err)
		}
	} else {
		conf, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, ferrors.Error(context.Background(), err)
	}
	return conf, nil
}

func SetupRemoteConfig(remoteConfPath string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if utils.DevMode() {
		configOverrides.CurrentContext = devModeRemoteContext
	}
	remoteKubeConf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	remoteConf, err := remoteKubeConf.ClientConfig()
	if err != nil {
		return nil, ferrors.Error(context.Background(), err)
	}
	return remoteConf, nil
}
