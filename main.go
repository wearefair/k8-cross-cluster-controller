package main

import (
	"context"
	"flag"
	"os"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/utils"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	EnvKubeConfigPath = "KUBECONFIG_PATH"
	// TODO: These should probably be configurable as flags
	devModeInternalContext = "prototype-general"
	devModeRemoteContext   = "prototype-secure"
)

var (
	kubeconfig string
	logger     = logging.Logger()
)

func main() {
	//	cmd.Execute()
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv(EnvKubeConfigPath), "Path to Kubeconfig")
	flag.Parse()
}

// If the controller is configured to run in development mode, it is configured to
// load the configuration from the default kubeconfig path ($HOME/.kube/config).
// Otherwise, it'll run the internal client with the in cluster config
// and the remote config from the config path that's passed in
func setupInternalConfig() (*rest.Config, error) {
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
