package k8

import (
	"context"

	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EndpointsReader struct {
	Events chan *EndpointsRequest
	client kubernetes.Interface
}

type EndpointsWriter struct {
	Events chan *EndpointsRequest
	client kubernetes.Interface
}

func NewEndpointsReader(clientset kubernetes.Interface) *EndpointsReader {
	return &EndpointsReader{
		Events: make(chan *EndpointsRequest),
		client: clientset,
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
	req := &EndpointsRequest{
		Type:      requestType,
		Endpoints: endpoints,
	}
	e.Events <- req
}

func (e *EndpointsReader) Client() kubernetes.Interface {
	return e.client
}

func NewEndpointsWriter(clientset kubernetes.Interface) *EndpointsWriter {
	return &EndpointsWriter{
		Events: make(chan *EndpointsRequest),
		client: clientset,
	}
}

func (e *EndpointsWriter) add(endpoints *v1.Endpoints) {
	_, err := e.client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Create(endpoints)
	if err != nil {
		errors.Error(context.Background(), err)
	}
}

func (e *EndpointsWriter) update(endpoints *v1.Endpoints) {
	_, err := e.client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Update(endpoints)
	if err != nil {
		errors.Error(context.Background(), err)
	}
}

func (e *EndpointsWriter) delete(endpoints *v1.Endpoints) {
	err := e.client.CoreV1().Endpoints(endpoints.ObjectMeta.Namespace).Delete(endpoints.Name, &metav1.DeleteOptions{})
	if err != nil {
		errors.Error(context.Background(), err)
	}
}

func (e *EndpointsWriter) Run() {
	for {
		request := <-e.Events
		switch request.Type {
		case RequestTypeAdd:
			e.add(request.Endpoints)
		case RequestTypeUpdate:
			e.update(request.Endpoints)
		case RequestTypeDelete:
			e.delete(request.Endpoints)
		}
	}
}
