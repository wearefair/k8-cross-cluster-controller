package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/controller"
	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	"github.com/wearefair/k8-cross-cluster-controller/pkg/utils"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	"github.com/wearefair/service-kit-go/uuid"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"

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

	// Run all writers
	go localEndpointsWriter.Run()
	go localServiceWriter.Run()

	// Run all readers/transformers
	logger.Info("Setting up transformers")
	go controller.ServiceAugmenter(localClient, remoteServiceReaderChan, intermediaryServiceReaderChan)
	go controller.EndpointsAugmenter(localClient, remoteEndpointsReaderChan, intermediaryEndpointsReaderChan)
	go controller.ServiceTransformer(intermediaryServiceReaderChan, localServiceWriterChan)
	go controller.EndpointsTransformer(intermediaryEndpointsReaderChan, localEndpointsWriterChan)

	stopChan := make(chan<- struct{})
	filter := func(options *metav1.ListOptions) {
		options.LabelSelector = k8.CrossClusterLabel
	}

	// Set up leader election callbacks
	run := func(stopChan <-chan struct{}) {
		// Watch remote endpoints and services
		logger.Info("Setting up watchers")
		k8.WatchEndpoints(remoteClient, remoteEndpointsReader, filter, stopChan)
		k8.WatchServices(remoteClient, remoteServiceReader, filter, stopChan)
		// Run service/endpoints cleaner
		logger.Info("Setting up service/endpoints cleaner")
		cleaner := controller.NewCleaner(localClient, remoteClient, localEndpointsWriterChan, localServiceWriterChan)
		go cleaner.Run(stopChan)
	}

	stop := func() {
		stopChan <- struct{}{}
	}

	id, err := os.Hostname()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Reference for setup:
	// https://github.com/kubernetes/kubernetes/blob/dce1b881284a103909f5cfa969ff56e5e0565362/cmd/cloud-controller-manager/app/controllermanager.go#L157-L190
	id = id + "_" + uuid.UUID()
	broadcaster := record.NewBroadcaster()
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{})
	lock, err := resourcelock.New(
		resourcelock.EndpointsResourceLock,
		metav1.NamespaceDefault,
		"cross-cluster-controller",
		localClient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		},
	)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Set up leader election
	logger.Info("Setting up leader election")
	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: run,
		OnStoppedLeading: stop,
	}

	config := leaderelection.LeaderElectionConfig{
		Callbacks:     callbacks,
		Lock:          lock,
		LeaseDuration: 1 * time.Minute,
		RenewDeadline: 30 * time.Second,
		RetryPeriod:   5 * time.Second,
	}

	leaderelection.RunOrDie(config)
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

// If a remote conf path is passed in, it will load it up with the explicit path. Otherwise
// it'll load the conf from the default kubeconfig path ($HOME/.kube/config).
// If it's run in dev mode, it will run in the context that's set
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
