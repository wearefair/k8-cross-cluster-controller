package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/cleaner"
	"github.com/wearefair/k8-cross-cluster-controller/pkg/controller"
	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
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
	EnvDevMode                  = "DEV_MODE"
	EnvKubeConfigPath           = "KUBECONFIG_PATH"
	fairSystemK8Namespace       = "fair-system"
	leaderElectionLeaseDuration = 1 * time.Minute
	leaderElectionRenewDeadline = 30 * time.Second
	leaderElectionRetryPeriod   = 5 * time.Second
)

var (
	devMode    string
	kubeconfig string
	// These are only set and used when the controller is running in dev mode
	localContext  string
	remoteContext string
	logger        = logging.Logger()
)

func main() {
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv(EnvKubeConfigPath), "Path to kubeconfig for remote cluster")
	flag.StringVar(&devMode, "devmode", os.Getenv(EnvDevMode), "Dev mode flag")
	flag.StringVar(&localContext, "local-context", "prototype-general", "DEV MODE: Context override for the local cluster. Defaults to prototype-general")
	flag.StringVar(&remoteContext, "remote-context", "prototype-secure", "DEV MODE: Context override for the remote cluster. Defaults to prototype-secure")
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

	logger.Info("Setting up local writers")
	localServiceWriterChan := make(chan *k8.ServiceRequest)
	localServiceWriter := k8.NewServiceWriter(localClient, localServiceWriterChan)
	localEndpointsWriterChan := make(chan *k8.EndpointsRequest)
	localEndpointsWriter := k8.NewEndpointsWriter(localClient, localEndpointsWriterChan)
	go localServiceWriter.Run()
	go localEndpointsWriter.Run()

	logger.Info("Setting up remote readers")
	remoteServiceReaderChan := make(chan *k8.ServiceRequest)
	remoteServiceReader := k8.NewServiceReader(remoteServiceReaderChan)
	remoteEndpointsReaderChan := make(chan *k8.EndpointsRequest)
	remoteEndpointsReader := k8.NewEndpointsReader(remoteEndpointsReaderChan)

	// Set up transformers
	logger.Info("Setting up transformers")
	augmenter := &controller.Augmenter{Client: localClient}
	go controller.EndpointsPipeline(remoteEndpointsReaderChan, localEndpointsWriterChan, augmenter.Endpoints, controller.EndpointsWhitelist)
	go controller.ServicePipeline(remoteServiceReaderChan, localServiceWriterChan, augmenter.Service, controller.ServiceWhitelist)

	filter := func(options *metav1.ListOptions) {
		options.LabelSelector = k8.CrossClusterLabel
	}

	logger.Info("Setting up service/endpoints cleaner")
	cleaner := cleaner.New(localClient, remoteClient, localEndpointsWriterChan, localServiceWriterChan)
	// Set up leader election callback funcs
	// Reference for leader election setup:
	// https://github.com/kubernetes/kubernetes/blob/dce1b881284a103909f5cfa969ff56e5e0565362/cmd/cloud-controller-manager/app/controllermanager.go#L157-L190
	run := func(stopChan <-chan struct{}) {
		logger.Info("Setting up watchers")
		k8.WatchEndpoints(remoteClient, remoteEndpointsReader, filter, stopChan)
		k8.WatchServices(remoteClient, remoteServiceReader, filter, stopChan)

		go cleaner.Run(stopChan, filter)
	}

	// Create a unique identifier based off of hostname and UUID
	id, err := os.Hostname()
	if err != nil {
		logger.Fatal(err.Error())
	}
	id = id + "_" + uuid.UUID()
	broadcaster := record.NewBroadcaster()
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{})
	lock, err := resourcelock.New(
		resourcelock.EndpointsResourceLock,
		fairSystemK8Namespace,
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
	logger.Info("Setting up leader election")
	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: run,
	}
	config := leaderelection.LeaderElectionConfig{
		Callbacks:     callbacks,
		Lock:          lock,
		LeaseDuration: leaderElectionLeaseDuration,
		RenewDeadline: leaderElectionRenewDeadline,
		RetryPeriod:   leaderElectionRetryPeriod,
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
	if devModeEnabled() {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{
			CurrentContext: localContext,
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
	if devModeEnabled() {
		configOverrides.CurrentContext = remoteContext
	}
	remoteKubeConf := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	remoteConf, err := remoteKubeConf.ClientConfig()
	if err != nil {
		return nil, ferrors.Error(context.Background(), err)
	}
	return remoteConf, nil
}

func devModeEnabled() bool {
	if devMode == "true" {
		return true
	}
	return false
}
