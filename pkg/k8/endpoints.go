package k8

import (
	"context"

	"go.uber.org/zap"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EndpointsReader struct {
	Events chan *EndpointsRequest
}

type EndpointsWriter struct {
	Events chan *EndpointsRequest
	Client kubernetes.Interface
}

func NewEndpointsReader(events chan *EndpointsRequest) *EndpointsReader {
	return &EndpointsReader{
		Events: events,
	}
}

func (e *EndpointsReader) Add(obj interface{}) {
	e.sendRequest(obj, RequestTypeAdd)
}

func (e *EndpointsReader) Update(_, newObj interface{}) {
	e.sendRequest(newObj, RequestTypeUpdate)
}

func (e *EndpointsReader) Delete(obj interface{}) {
	e.sendRequest(obj, RequestTypeDelete)
}

func (e *EndpointsReader) sendRequest(obj interface{}, requestType RequestType) {
	endpoints := obj.(*v1.Endpoints)
	logger.Info("Sending endpoints request", zap.String("requestType", RequestTypeMap[requestType]),
		zap.String("name", endpoints.Name), zap.String("namespace", endpoints.ObjectMeta.Namespace))
	req := &EndpointsRequest{
		Type:            requestType,
		RemoteEndpoints: endpoints,
	}
	e.Events <- req
}

func NewEndpointsWriter(clientset kubernetes.Interface, events chan *EndpointsRequest) *EndpointsWriter {
	return &EndpointsWriter{
		Events: events,
		Client: clientset,
	}
}

func (e *EndpointsWriter) add(endpoints *v1.Endpoints) {
	logger.Info("Creating endpoints", zap.String("name", endpoints.Name),
		zap.String("namespace", endpoints.ObjectMeta.Namespace))
	e.create(endpoints)
}

func (e *EndpointsWriter) update(endpoints *v1.Endpoints) {
	ctx := context.Background()
	update := func() error {
		logger.Info("Updating endpoints", zap.String("name", endpoints.Name),
			zap.String("namespace", endpoints.ObjectMeta.Namespace))
		_, err := e.Client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Update(endpoints)
		if err != nil {
			// If the endpoint doesn't exist, attempt to create it
			if ResourceNotExist(err) {
				e.create(endpoints)
			} else {
				return err
			}
		}
		return nil
	}
	exponentialBackOff(ctx, update)
}

func (e *EndpointsWriter) create(endpoints *v1.Endpoints) {
	ctx := context.Background()
	create := func() error {
		logger.Info("Creating endpoints", zap.String("name", endpoints.Name),
			zap.String("namespace", endpoints.ObjectMeta.Namespace))
		_, err := e.Client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Create(endpoints)
		if err != nil {
			return err
		}
		return nil
	}
	exponentialBackOff(ctx, create)
}

func (e *EndpointsWriter) delete(endpoints *v1.Endpoints) {
	ctx := context.Background()
	delete := func() error {
		logger.Info("Deleting endpoints", zap.String("name", endpoints.Name),
			zap.String("namespace", endpoints.ObjectMeta.Namespace))
		err := e.Client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Delete(endpoints.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	}
	exponentialBackOff(ctx, delete)
}

func (e *EndpointsWriter) Run() {
	for {
		request := <-e.Events
		switch request.Type {
		case RequestTypeAdd:
			e.add(request.LocalEndpoints)
		case RequestTypeUpdate:
			e.update(request.LocalEndpoints)
		case RequestTypeDelete:
			e.delete(request.LocalEndpoints)
		}
	}
}
