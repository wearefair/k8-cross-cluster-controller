package controller

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/wearefair/k8-cross-cluster-controller/k8"
	"github.com/wearefair/service-kit-go/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	EnvConfigPath = "CONFIG_PATH"
)

type Controller struct {
	RequestChan    chan *k8.ServiceRequest
	InternalClient *k8.Client
	RemoteClient   *k8.Client
}

func NewController(clusterName, host, token string) (*Controller, error) {
	requestChan := make(chan *k8.ServiceRequest)
	internal, err := k8.NewClient(nil, requestChan)
	if err != nil {
		return nil, err
	}
	remoteConf := &k8.ClusterConfig{
		ClusterName: clusterName,
		Host:        host,
		Token:       token,
	}
	remote, err := k8.NewClient(remoteConf, requestChan)
	if err != nil {
		return nil, err
	}
	controller := &Controller{
		RequestChan:    requestChan,
		InternalClient: internal,
		RemoteClient:   remote,
	}
	return controller, nil
}

func Coordinate() error {
	requestChan := make(chan *k8.ServiceRequest)
	internal, err := k8.NewClient(nil, requestChan)
	if err != nil {
		return err
	}
	confPath := os.Getenv(EnvConfigPath)
	remoteConf, err := config(confPath)
	if err != nil {
		return err
	}
	remote, err := k8.NewClient(remoteConf, requestChan)
	if err != nil {
		return err
	}
	if err := k8.WatchServices(remote); err != nil {
		return err
	}
	for {
		request := <-requestChan
		switch request.Type {
		case k8.AddService:
			err = internal.CreateService(request)
		case k8.UpdateService:
			err = internal.UpdateService(request)
		case k8.DeleteService:
			err = internal.DeleteService(request)
		default:
			// How did you even get here?
			break
		}
		if err != nil {
			fmt.Println(err)
		}
	}
}

func config(path string) (*k8.ClusterConfig, error) {
	conf := &k8.ClusterConfig{}
	ctx := context.Background()
	confPath, err := pathHelper(path)
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
