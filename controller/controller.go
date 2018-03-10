package controller

import "github.com/wearefair/k8-cross-cluster-controller/k8"

type Controller struct {
	RequestChan    chan *k8.ServiceRequest
	InternalClient *k8.Client
	RemoteClient   *k8.Client
}

func NewController(clusterName, host, token string) (*Controller, error) {
	requestChan := make(chan *ServiceRequest)
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

// Might not need the controller object
func Coordinate() error {
	requestChan := make(chan *ServiceRequest)
	internal, err := k8.NewClient(nil, requestChan)
	if err != nil {
		return err
	}
	remoteConf := &k8.ClusterConfig{
		ClusterName: clusterName,
		Host:        host,
		Token:       token,
	}
	remote, err := k8.NewClient(remoteConf, requestChan)
	if err != nil {
		return err
	}
	for {
		request := <-requestChan
		select {
		case request.Type == k8.AddService:
		case request.Type == k8.UpdateService:
		case request.Type == k8.DeleteService:
		default:
			// How did you even get here?
			break
		}
	}
}
