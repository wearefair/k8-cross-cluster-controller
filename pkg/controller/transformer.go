package controller

import (
	"context"

	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type serviceTransformer func(remoteSvc, localSvc *v1.Service) error

type endpointsTransformer func(remoteEndpoints, localEndpoints *v1.Endpoints) error

type EndpointsUpdater struct {
	client kubernetes.Interface
}

func (e *EndpointsUpdater) Transform(remoteEndpoints, localEndpoints *v1.Endpoints) (err error) {
	localEndpoints, err = e.client.CoreV1().Endpoints(remoteEndpoints.ObjectMeta.Namespace).Get(remoteEndpoints.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}

type ServiceUpdater struct {
	client kubernetes.Interface
}

func (s *ServiceUpdater) Transform(remoteSvc, localSvc *v1.Service) (err error) {
	localSvc, err = s.client.CoreV1().Services(remoteSvc.ObjectMeta.Namespace).Get(remoteSvc.Name, metav1.GetOptions{})
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}
