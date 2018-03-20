package k8

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type EndpointsReader struct {
	Events chan *EndpointsRequest
	Client kubernetes.Interface
}

type EndpointsWriter struct {
	Events chan *EndpointsRequest
	Client kubernetes.Interface
}

func NewEndpointsReader(clientset kubernetes.Interface) *EndpointsReader {
	return &EndpointsReader{
		Events: make(chan *EndpointsRequest),
		Client: clientset,
	}
}

func (e *EndpointsReader) Add(obj interface{}) {}

func (e *EndpointsReader) Update(oldObj, newObj interface{}) {}

func (e *EndpointsReader) Delete(obj interface{}) {}

func NewEndpointsWriter(clientset kubernetes.Interface) *EndpointsWriter {
	return &EndpointsWriter{
		Events: make(chan *EndpointsRequest),
		Client: clientset,
	}
}

func (e *EndpointsWriter) add(endpoints *v1.Endpoints) {}

func (e *EndpointsWriter) update(endpoints *v1.Endpoints) {}

func (e *EndpointsWriter) delete(endpoints *v1.Endpoints) {}

func (e *EndpointsWriter) Run() {}
