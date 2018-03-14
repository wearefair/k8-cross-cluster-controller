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
	K8Client      kubernetes.Interface
	ServiceChan   chan *ServiceRequest
	EndpointsChan chan *EndpointsRequest
}

func NewRemoteClient(clientset kubernetes.Interface, serviceChan chan *ServiceRequest, endpointsChan chan *EndpointsRequest) *RemoteClient {
	return &RemoteClient{
		K8Client:      clientset,
		ServiceChan:   serviceChan,
		EndpointsChan: endpointsChan,
	}
}

func (r *RemoteClient) WatchAddService(obj interface{}) {
	svc := obj.(*v1.Service)
	logger.Info("Service added", zap.String("name", svc.ObjectMeta.Name))
	r.sendServiceRequest(svc, RequestTypeAdd)
}

func (r *RemoteClient) WatchDeleteService(obj interface{}) {
	svc := obj.(*v1.Service)
	logger.Info("Service deleted", zap.String("name", svc.ObjectMeta.Name))
	r.sendServiceRequest(svc, RequestTypeDelete)
}

func (r *RemoteClient) WatchUpdateService(_, new interface{}) {
	svc := new.(*v1.Service)
	logger.Info("Service updated", zap.String("name", svc.ObjectMeta.Name))
	r.sendServiceRequest(svc, RequestTypeUpdate)
}

func (r *RemoteClient) sendServiceRequest(svc *v1.Service, requestType RequestType) {
	req := &ServiceRequest{
		Type:    requestType,
		Service: svc,
	}
	r.ServiceChan <- req
}

func (r *RemoteClient) WatchAddEndpoints(obj interface{}) {
	endpoints := obj.(*v1.Endpoints)
	logger.Info("Endpoints added", zap.String("name", endpoints.ObjectMeta.Name))
	r.sendEndpointsRequest(endpoints, RequestTypeAdd)
}

func (r *RemoteClient) WatchUpdateEndpoints(_, new interface{}) {
	endpoints := new.(*v1.Endpoints)
	logger.Info("Endpoints updated", zap.String("name", endpoints.ObjectMeta.Name))
	r.sendEndpointsRequest(endpoints, RequestTypeUpdate)
}

func (r *RemoteClient) WatchDeleteEndpoints(obj interface{}) {
	endpoints := obj.(*v1.Endpoints)
	logger.Info("Endpoints deleted", zap.String("name", endpoints.ObjectMeta.Name))
	r.sendEndpointsRequest(endpoints, RequestTypeDelete)
}

func (r *RemoteClient) sendEndpointsRequest(endpoints *v1.Endpoints, requestType RequestType) {
	req := &EndpointsRequest{
		Type:      requestType,
		Endpoints: endpoints,
	}
	r.EndpointsChan <- req
}

func (r *RemoteClient) Client() kubernetes.Interface {
	return r.K8Client
}

func (r *RemoteClient) getEndpointsFromService(svc *v1.Service) (*v1.Endpoints, error) {
	name := svc.ObjectMeta.Name
	endpoints, err := r.K8Client.CoreV1().Endpoints(defaultNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Error(context.Background(), err)
	}
	return endpoints, nil
}
