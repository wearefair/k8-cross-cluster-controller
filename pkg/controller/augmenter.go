package controller

import (
	"context"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// The Augmenter "augments" an EndpointsRequest or ServiceRequest with a local K8 resource
// that will further be transformed down the pipeline before the writers act upon it.
type Augmenter struct {
	Client kubernetes.Interface
}

func (a *Augmenter) Service(req *k8.ServiceRequest) error {
	switch req.Type {
	case k8.RequestTypeAdd:
		req.LocalService = &v1.Service{}
	case k8.RequestTypeUpdate:
		localService, err := a.Client.CoreV1().Services(req.RemoteService.ObjectMeta.Namespace).Get(req.RemoteService.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Error(context.Background(), err)
		}
		req.LocalService = localService
	// We don't apply any transformations on the local endpoint, because a delete just requires the K8 object name
	case k8.RequestTypeDelete:
		req.LocalService = req.RemoteService
	}
	return nil
}

func (a *Augmenter) Endpoints(req *k8.EndpointsRequest) error {
	switch req.Type {
	case k8.RequestTypeAdd:
		req.LocalEndpoints = &v1.Endpoints{}
	case k8.RequestTypeUpdate:
		localEndpoints, err := a.Client.CoreV1().Endpoints(req.RemoteEndpoints.ObjectMeta.Namespace).Get(req.RemoteEndpoints.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Error(context.Background(), err)
		}
		req.LocalEndpoints = localEndpoints
	// We don't apply any transformations on the local endpoint, because a delete just requires the K8 object name
	case k8.RequestTypeDelete:
		req.LocalEndpoints = req.RemoteEndpoints
	}
	return nil
}