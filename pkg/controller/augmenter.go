package controller

import (
	"context"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Augmenter struct {
	Client kubernetes.Interface
}

func (a *Augmenter) Service(req *k8.ServiceRequest) error {
	req.LocalService = &v1.Service{}
	switch req.Type {
	case k8.RequestTypeUpdate:
		localService, err := a.Client.CoreV1().Services(req.RemoteService.ObjectMeta.Namespace).Get(req.RemoteService.Name, metav1.GetOptions{})
		if err != nil {
			// If the resource doesn't exist, transform the request into an add to create it
			if k8.ResourceNotExist(err) {
				req.Type = k8.RequestTypeAdd
				return nil
			} else {
				return errors.Error(context.Background(), err)
			}
		}
		req.LocalService = localService
	// We can just copy the remote service, because a delete just requires the K8 resource name
	case k8.RequestTypeDelete:
		req.LocalService = req.RemoteService
	}
	return nil
}

func (a *Augmenter) Endpoints(req *k8.EndpointsRequest) error {
	req.LocalEndpoints = &v1.Endpoints{}
	switch req.Type {
	case k8.RequestTypeUpdate:
		localEndpoints, err := a.Client.CoreV1().Endpoints(req.RemoteEndpoints.ObjectMeta.Namespace).Get(req.RemoteEndpoints.Name, metav1.GetOptions{})
		if err != nil {
			// If the resource doesn't exist, transform the request into an add to create it
			if k8.ResourceNotExist(err) {
				req.Type = k8.RequestTypeAdd
				return nil
			} else {
				return errors.Error(context.Background(), err)
			}
		}
		req.LocalEndpoints = localEndpoints
	// We can just copy the remote endpoint, because a delete just requires the K8 resource name
	case k8.RequestTypeDelete:
		req.LocalEndpoints = req.RemoteEndpoints
	}
	return nil
}
