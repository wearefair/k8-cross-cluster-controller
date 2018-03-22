package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/controller"
	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	"github.com/wearefair/k8-cross-cluster-controller/pkg/utils"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	logger.Info("Setting up local K8 client")
	localClient, err := kubernetes.NewForConfig(localConf)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Setting up remote K8 client")
	remoteClient, err := kubernetes.NewForConfig(remoteConf)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Set up local readers and writers
	logger.Info("Setting up local readers and writers")

	localServiceWriterChan := make(chan *k8.ServiceRequest)
	localEndpointsWriterChan := make(chan *k8.EndpointsRequest)
	localEndpointsWriter := k8.NewEndpointsWriter(localClient, localEndpointsWriterChan)
	localServiceWriter := k8.NewServiceWriter(localClient, localServiceWriterChan)

	// Set up remote readers and writers
	logger.Info("Setting up remote readers")
	remoteServiceReaderChan := make(chan *k8.ServiceRequest)
	remoteEndpointsReaderChan := make(chan *k8.EndpointsRequest)
	remoteServiceReader := k8.NewServiceReader(remoteServiceReaderChan)
	remoteEndpointsReader := k8.NewEndpointsReader(remoteEndpointsReaderChan)

	// Create reader chans that aren't attached to readers hooked up to
	intermediaryServiceReaderChan := make(chan *k8.ServiceRequest)
	intermediaryEndpointsReaderChan := make(chan *k8.EndpointsRequest)

	stopChan := make(chan struct{})
	filter := func(options *metav1.ListOptions) {
		options.LabelSelector = k8.CrossClusterLabel
	}

	// Watch remote endpoints and services
	logger.Info("Setting up watchers")
	k8.WatchEndpoints(remoteClient, remoteEndpointsReader, filter, stopChan)
	k8.WatchServices(remoteClient, remoteServiceReader, filter, stopChan)

	// Run all writers
	go localEndpointsWriter.Run()
	go localServiceWriter.Run()

	// Run all readers/transformers
	logger.Info("Setting up transformers")
	go controller.ServiceAugmenter(localClient, remoteServiceReaderChan, intermediaryServiceReaderChan)
	go controller.EndpointsAugmenter(localClient, remoteEndpointsReaderChan, intermediaryEndpointsReaderChan)
	go controller.ServiceTransformer(intermediaryServiceReaderChan, localServiceWriterChan)
	go controller.EndpointsTransformer(intermediaryEndpointsReaderChan, localEndpointsWriterChan)

	// Setting up cleaner
	logger.Info("Setting up service/endpoints cleaner")
	cleaner := controller.NewCleaner(localClient, remoteClient, localEndpointsWriterChan, localServiceWriterChan)
	go cleaner.Run()

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
	if remoteConfPath != "" {
		loadingRules.ExplicitPath = remoteConfPath
	}
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
