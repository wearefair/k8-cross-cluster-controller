package k8

import (
	"context"

	"go.uber.org/zap"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/cenkalti/backoff"
	"k8s.io/apimachinery/pkg/api/errors"
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
	update := func() error {
		logger.Info("Updating endpoints", zap.String("name", endpoints.Name),
			zap.String("namespace", endpoints.ObjectMeta.Namespace))
		_, err := e.Client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Update(endpoints)
		if err != nil {
			// If the endpoint doesn't exist, attempt to create it
			if ResourceNotExist(err) {
				e.create(endpoints)
				return nil
			}
			if PermanentError(err) {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}
	exponentialBackOff(context.Background(), update)
}

func (e *EndpointsWriter) create(endpoints *v1.Endpoints) {
	create := func() error {
		logger.Info("Creating endpoints", zap.String("name", endpoints.Name),
			zap.String("namespace", endpoints.ObjectMeta.Namespace))
		_, err := e.Client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Create(endpoints)
		if err != nil {
			// If the resource already exists, we should not attempt backoff behavior
			if errors.IsAlreadyExists(err) {
				logger.Info("Endpoints already exists, skipping create",
					zap.String("name", endpoints.Name),
					zap.String("namespace", endpoints.ObjectMeta.Namespace))
				return nil
			}
			return err
		}
		return nil
	}
	exponentialBackOff(context.Background(), create)
}

func (e *EndpointsWriter) delete(endpoints *v1.Endpoints) {
	delete := func() error {
		logger.Info("Deleting endpoints", zap.String("name", endpoints.Name),
			zap.String("namespace", endpoints.ObjectMeta.Namespace))
		err := e.Client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Delete(endpoints.Name, &metav1.DeleteOptions{})
		if err != nil {
			// If resource does not exist, we should not attempt backoff behavior
			if PermanentError(err) {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}
	exponentialBackOff(context.Background(), delete)
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
