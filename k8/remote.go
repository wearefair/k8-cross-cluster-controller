package k8

import (
	"context"

	"github.com/wearefair/service-kit-go/errors"
	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RemoteClient struct {
	K8Client    kubernetes.Interface
	RequestChan chan *ServiceRequest
}

func NewRemoteClient(clientset kubernetes.Interface, requestChan chan *ServiceRequest) *RemoteClient {
	return &RemoteClient{
		K8Client:    clientset,
		RequestChan: requestChan,
	}
}

func (r *RemoteClient) AddService(obj interface{}) {
	svc := obj.(*v1.Service)
	logger.Info("Service added", zap.String("name", svc.ObjectMeta.Name))
	endpoint, err := r.getEndpointsFromService(svc)
	if err != nil {
		return
	}
	req := r.createServiceRequest(endpoint, AddService)
	r.RequestChan <- req
}

func (r *RemoteClient) DeleteService(obj interface{}) {
	svc := obj.(*v1.Service)
	logger.Info("Service deleted", zap.String("name", svc.ObjectMeta.Name))
	endpoint, err := r.getEndpointsFromService(svc)
	if err != nil {
		return
	}
	req := r.createServiceRequest(endpoint, DeleteService)
	r.RequestChan <- req
}

func (r *RemoteClient) UpdateService(old, new interface{}) {
	logger.Info("Got update service event")
	svc := new.(*v1.Service)
	endpoint, err := r.getEndpointsFromService(svc)
	if err != nil {
		return
	}
	req := r.createServiceRequest(endpoint, UpdateService)
	r.RequestChan <- req
}

func (r *RemoteClient) getEndpointsFromService(svc *v1.Service) (*v1.Endpoints, error) {
	name := svc.ObjectMeta.Name
	endpoints, err := r.K8Client.CoreV1().Endpoints(defaultNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Error(context.Background(), err)
	}
	return endpoints, nil
}

func (r *RemoteClient) createServiceRequest(endpoint *v1.Endpoints, requestType ServiceRequestType) *ServiceRequest {
	return &ServiceRequest{
		Type:     requestType,
		Endpoint: endpoint,
	}
}
