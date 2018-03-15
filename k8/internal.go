package k8

import (
	"context"

	"go.uber.org/zap"

	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type InternalClient struct {
	Cluster             string
	K8Client            kubernetes.Interface
	InternalServiceChan chan *ServiceRequest
	RemoteServiceChan   chan *ServiceRequest
	RemoteEndpointsChan chan *EndpointsRequest
}

func NewInternalClient(clientset kubernetes.Interface, internalServiceChan, remoteServiceChan chan *ServiceRequest, remoteEndpointsChan chan *EndpointsRequest) *InternalClient {
	return &InternalClient{
		K8Client:            clientset,
		InternalServiceChan: internalServiceChan,
		RemoteServiceChan:   remoteServiceChan,
		RemoteEndpointsChan: remoteEndpointsChan,
	}
}

func (k *InternalClient) WatchAddService(obj interface{}) {
	svc := obj.(*v1.Service)
	logger.Info("Service added", zap.String("name", svc.ObjectMeta.Name))
	k.sendServiceRequest(svc, RequestTypeAdd)
}

func (k *InternalClient) WatchUpdateService(_, new interface{}) {
	svc := new.(*v1.Service)
	logger.Info("Service updated", zap.String("name", svc.ObjectMeta.Name))
	k.sendServiceRequest(svc, RequestTypeUpdate)
}

func (k *InternalClient) WatchDeleteService(obj interface{}) {
	svc := obj.(*v1.Service)
	logger.Info("Service deleted", zap.String("name", svc.ObjectMeta.Name))
	k.sendServiceRequest(svc, RequestTypeDelete)
}

func (k *InternalClient) sendServiceRequest(svc *v1.Service, requestType RequestType) {
	req := &ServiceRequest{
		Type:    requestType,
		Service: svc,
	}
	k.InternalServiceChan <- req
}

func (k *InternalClient) HandleRemoteAddService(request *ServiceRequest) error {
	_, err := k.K8Client.CoreV1().Services(defaultNamespace).Create(
		SanitizeService(request.Service.DeepCopy()),
	)
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}

func (k *InternalClient) HandleRemoteDeleteService(request *ServiceRequest) error {
	err := k.K8Client.CoreV1().Services(defaultNamespace).Delete(request.Service.ObjectMeta.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}

func (k *InternalClient) HandleRemoteUpdateService(request *ServiceRequest) error {
	svc := request.Service.DeepCopy()
	svc.Spec.ClusterIP = ""
	_, err := k.K8Client.CoreV1().Services(defaultNamespace).Update(svc)
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}

// TODO: Must strip out resource version
func (k *InternalClient) HandleRemoteAddEndpoints(request *EndpointsRequest) error {
	_, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Create(
		SanitizeEndpoints(request.Endpoints.DeepCopy()),
	)
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}

func (k *InternalClient) HandleRemoteUpdateEndpoints(request *EndpointsRequest) error {
	_, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Update(
		SanitizeEndpoints(request.Endpoints.DeepCopy()),
	)
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}

func (k *InternalClient) HandleRemoteDeleteEndpoints(request *EndpointsRequest) error {
	err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Delete(request.Endpoints.ObjectMeta.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}

// HandleInternalAddService takes a service request, looks for an endpoint of the same name, and applies the controller
// label to it
func (k *InternalClient) HandleInternalAddService(request *ServiceRequest) error {
	return k.applyCrossClusterLabelToEndpoints(request.Service)
}

func (k *InternalClient) HandleInternalUpdateService(request *ServiceRequest) error {
	return k.applyCrossClusterLabelToEndpoints(request.Service)
}

func (k *InternalClient) applyCrossClusterLabelToEndpoints(svc *v1.Service) error {
	endpoints, err := k.getEndpointsFromService(svc)
	if err != nil {
		return err
	}
	labels := endpoints.ObjectMeta.GetLabels()
	// If the cross cluster label doesn't exist, add it to the endpoints
	if _, ok := labels[CrossClusterServiceLabelKey]; !ok {
		labels[CrossClusterServiceLabelKey] = CrossClusterServiceLabelValue
		endpoints.ObjectMeta.SetLabels(labels)
		if _, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Update(endpoints); err != nil {
			return errors.Error(context.Background(), err)
		}
	}
	return nil
}

func (k *InternalClient) HandleInternalServiceEvents() error {
	for {
		var err error
		request := <-k.InternalServiceChan
		switch request.Type {
		case RequestTypeAdd:
			err = k.HandleInternalAddService(request)
		case RequestTypeUpdate:
			err = k.HandleInternalUpdateService(request)
		case RequestTypeDelete:
			logger.Info("Got delete request type, ignoring")
		default:
			logger.Error("Got impossible request type")
		}
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

func (k *InternalClient) HandleRemoteServiceEvents() {
	for {
		var err error
		request := <-k.RemoteServiceChan
		switch request.Type {
		case RequestTypeAdd:
			err = k.HandleRemoteAddService(request)
		case RequestTypeUpdate:
			err = k.HandleRemoteUpdateService(request)
		case RequestTypeDelete:
			err = k.HandleRemoteDeleteService(request)
		default:
			logger.Error("Got impossible request type")
		}
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

func (k *InternalClient) HandleRemoteEndpointsEvents() {
	for {
		var err error
		request := <-k.RemoteEndpointsChan
		switch request.Type {
		case RequestTypeAdd:
			err = k.HandleRemoteAddEndpoints(request)
		case RequestTypeUpdate:
			err = k.HandleRemoteUpdateEndpoints(request)
		case RequestTypeDelete:
			err = k.HandleRemoteDeleteEndpoints(request)
		default:
			logger.Error("Got impossible request type")
		}
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

func (k *InternalClient) Client() kubernetes.Interface {
	return k.K8Client
}

func (k *InternalClient) getEndpointsFromService(svc *v1.Service) (*v1.Endpoints, error) {
	name := svc.ObjectMeta.Name
	endpoints, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Error(context.Background(), err)
	}
	return endpoints, nil
}