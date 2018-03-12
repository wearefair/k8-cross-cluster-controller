package controller

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/wearefair/k8-cross-cluster-controller/k8"
	"github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	yaml "gopkg.in/yaml.v2"
)

const (
	EnvConfigPath = "CONFIG_PATH"
)

var (
	logger = logging.Logger()
)

type Controller struct {
	RequestChan    chan *k8.ServiceRequest
	InternalClient *k8.InternalClient
	RemoteClient   *k8.RemoteClient
}

func Coordinate() error {
	requestChan := make(chan *k8.ServiceRequest)
	logger.Info("Setting up local client")
	internal, err := k8.NewClient(nil, requestChan)
	if err != nil {
		return err
	}
	confPath := os.Getenv(EnvConfigPath)
	remoteConf, err := config(confPath)
	if err != nil {
		return err
	}
	logger.Info("Setting up remote client")
	remote, err := k8.NewClient(remoteConf, requestChan)
	if err != nil {
		return err
	}
	logger.Info("Watching services")
	if err := k8.WatchServices(remote); err != nil {
		return err
	}
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
			// How did you even get here?
			break
		}
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

func config(path string) (*k8.ClusterConfig, error) {
	conf := &k8.ClusterConfig{}
	ctx := context.Background()
	confPath, err := pathHelper(path)
	logger.Info("Config path", zap.String("path", confPath))
	if err != nil {
		return nil, err
	}
	confFile, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, errors.Error(ctx, err)
	}
	if err := yaml.Unmarshal(confFile, conf); err != nil {
		return nil, errors.Error(ctx, err)
	}
	return conf, nil
}

func pathHelper(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	finalPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", errors.Error(context.Background(), err)
	}
	return finalPath, nil
}
