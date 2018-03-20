package k8

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceReader struct {
	Events chan *ServiceRequest
	Client kubernetes.Interface
}

type ServiceWriter struct {
	Events chan *ServiceRequest
	Client kubernetes.Interface
}

func NewServiceReader(clientset kubernetes.Interface) *ServiceReader {
	return &ServiceReader{
		Events: make(chan *ServiceRequest),
		Client: clientset,
	}
}

func (s *ServiceReader) Add(obj interface{}) {}

func (s *ServiceReader) Update(oldObj, newObj interface{}) {}

func (s *ServiceReader) Delete(obj interface{}) {}

func NewServiceWriter(clientset kubernetes.Interface) *ServiceWriter {
	return &ServiceWriter{
		Events: make(chan *ServiceRequest),
		Client: clientset,
	}
}

func (s *ServiceWriter) add(svc *v1.Service) {}

func (s *ServiceWriter) update(svc *v1.Service) {}

func (s *ServiceWriter) delete(svc *v1.Service) {}

func (s *ServiceWriter) Run() {}
