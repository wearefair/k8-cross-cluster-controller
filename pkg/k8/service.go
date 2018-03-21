package k8

import (
	"context"

	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceReader struct {
	Events chan *ServiceRequest
	client kubernetes.Interface
}

type ServiceWriter struct {
	Events chan *ServiceRequest
	client kubernetes.Interface
}

func NewServiceReader(clientset kubernetes.Interface) *ServiceReader {
	return &ServiceReader{
		Events: make(chan *ServiceRequest),
		client: clientset,
	}
}

func (s *ServiceReader) Add(obj interface{}) {
	s.sendRequest(obj, RequestTypeAdd)
}

func (s *ServiceReader) Update(_, newObj interface{}) {
	s.sendRequest(newObj, RequestTypeUpdate)
}

func (s *ServiceReader) Delete(obj interface{}) {
	s.sendRequest(obj, RequestTypeDelete)
}

func (s *ServiceReader) sendRequest(obj interface{}, requestType RequestType) {
	service := obj.(*v1.Service)
	req := &ServiceRequest{
		Type:    requestType,
		Service: service,
	}
	s.Events <- req
}

func (s *ServiceReader) Client() kubernetes.Interface {
	return s.client
}

func NewServiceWriter(clientset kubernetes.Interface) *ServiceWriter {
	return &ServiceWriter{
		Events: make(chan *ServiceRequest),
		client: clientset,
	}
}

func (s *ServiceWriter) add(svc *v1.Service) {
	_, err := s.client.CoreV1().Services(svc.ObjectMeta.Namespace).Create(svc)
	if err != nil {
		errors.Error(context.Background(), err)
	}
}

func (s *ServiceWriter) update(svc *v1.Service) {
	_, err := s.client.CoreV1().Services(svc.ObjectMeta.Namespace).Update(svc)
	if err != nil {
		errors.Error(context.Background(), err)
	}
}

func (s *ServiceWriter) delete(svc *v1.Service) {
	err := s.client.CoreV1().Services(svc.ObjectMeta.Namespace).Delete(svc.Name, &metav1.DeleteOptions{})
	if err != nil {
		errors.Error(context.Background(), err)
	}
}

func (s *ServiceWriter) Run() {
	for {
		request := <-s.Events
		switch request.Type {
		case RequestTypeAdd:
			s.add(request.Service)
		case RequestTypeUpdate:
			s.update(request.Service)
		case RequestTypeDelete:
			s.delete(request.Service)
		}
	}
}
