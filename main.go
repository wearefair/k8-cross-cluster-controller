package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	"github.com/wearefair/k8-cross-cluster-controller/pkg/utils"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	EnvKubeConfigPath = "KUBECONFIG_PATH"
	// TODO: These should probably be configurable as flags
	devModeLocalContext  = "prototype-general"
	devModeRemoteContext = "prototype-secure"
)

var (
	kubeconfig string
	logger     = logging.Logger()
)

func main() {
	//	cmd.Execute()
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv(EnvKubeConfigPath), "Path to Kubeconfig")
	flag.Parse()

	localConf, err := setupLocalConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}
	remoteConf, err := setupRemoteConfig(kubeconfig)
	if err != nil {
		logger.Fatal(err.Error())
	}

	localClient, err := kubernetes.NewForConfig(localConf)
	if err != nil {
		logger.Fatal(err.Error())
	}

	remoteClient, err := kubernetes.NewForConfig(remoteConf)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Set up local readers and writers
	localServiceReader := k8.NewServiceReader(localClient)
	localEndpointsWriter := k8.NewEndpointsWriter(localClient)
	localServiceWriter := k8.NewServiceWriter(localClient)

	// Set up remote readers and writers
	remoteServiceReader := k8.NewServiceReader(remoteClient)
	remoteEndpointsReader := k8.NewEndpointsReader(remoteClient)

	stopChan := make(chan struct{})
	filter := func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", k8.CrossClusterServiceLabelKey, k8.CrossClusterServiceLabelValue)
	}

	// Watch local services
	k8.WatchServices(localServiceReader, filter, stopChan)

	// Watch remote endpoints and services
	k8.WatchEndpoints(remoteServiceReader, filter, stopChan)
	k8.WatchServices(remoteEndpointsReader, filter, stopChan)

	// Run all writers
	go localEndpointsWriter.Run()
	go localServiceWriter.Run()

	// Run all coordinators
	go LocalCoordinator(localClient, localServiceReader.Events, localEndpointsWriter.Events)
	go RemoteCoordinator(remoteServiceReader.Events, localServiceWriter.Events, remoteEndpointsReader.Events, localEndpointsWriter.Events)

	// Terminate watchers on SIGINT
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	logger.Info("Shutting down watchers")
	close(stopChan)
}

// If the controller is configured to run in development mode, it is configured to
// load the configuration from the default kubeconfig path ($HOME/.kube/config).
// Otherwise, it'll run the local client with the in cluster config
// and the remote config from the config path that's passed in
func setupLocalConfig() (*rest.Config, error) {
	var conf *rest.Config
	var err error
	if utils.DevMode() {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{
			CurrentContext: devModeLocalContext,
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

// TODO: If remoteConfPath is not an empty string, set it in the config overrides
func setupRemoteConfig(remoteConfPath string) (*rest.Config, error) {
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
