package controller

import (
	"context"

	"errors"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/wearefair/k8-cross-cluster-controller/k8"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	logger = logging.Logger()
)

func Coordinate(remoteConfPath string) error {
	ctx := context.Background()
	requestChan := make(chan *k8.ServiceRequest)

	logger.Info("Setting up internal client")
	internalConf, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	internalClient, err := kubernetes.NewForConfig(internalConf)
	if err != nil {
		return ferrors.Error(ctx, err)
	}
	internal := k8.NewInternalClient(internalClient, requestChan)

	logger.Info("Setting up remote client")
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if remoteConfPath != "" {
		loadingRules.ExplicitPath = remoteConfPath
	}
	remoteKubeConf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	remoteConf, err := remoteKubeConf.ClientConfig()
	if err != nil {
		return ferrors.Error(ctx, err)
	}
	remoteClient, err := kubernetes.NewForConfig(remoteConf)
	if err != nil {
		return ferrors.Error(ctx, err)
	}
	remote := k8.NewRemoteClient(remoteClient, requestChan)
	return coordinate(internal, remote, requestChan)
}

func coordinate(internal *k8.InternalClient, remote *k8.RemoteClient, requestChan chan *k8.ServiceRequest) error {
	logger.Info("Watching services")
	if err := k8.WatchServices(remote); err != nil {
		return err
	}

	var err error
	for {
		request := <-requestChan
		switch request.Type {
		case k8.AddService:
			logger.Info("Got add service request", zap.String("name", request.Endpoint.ObjectMeta.Name))
			err = internal.CreateService(request)
		case k8.UpdateService:
			logger.Info("Got update service request", zap.String("name", request.Endpoint.ObjectMeta.Name))
			err = internal.UpdateService(request)
		case k8.DeleteService:
			logger.Info("Got delete service request", zap.String("name", request.Endpoint.ObjectMeta.Name))
			err = internal.DeleteService(request)
		default:
			err = errors.New("Got impossible request type")
		}
		if err != nil {
			return ferrors.Error(context.Background(), err)
		}
	}
}
