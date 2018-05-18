package k8

import (
	"context"

	"github.com/cenkalti/backoff"
	"go.uber.org/zap"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceReader struct {
	Events chan *ServiceRequest
}

type ServiceWriter struct {
	Events chan *ServiceRequest
	Client kubernetes.Interface
}

func NewServiceReader(events chan *ServiceRequest) *ServiceReader {
	return &ServiceReader{
		Events: events,
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
	logger.Info("Sending service request", zap.String("requestType", RequestTypeMap[requestType]),
		zap.String("name", service.Name), zap.String("namespace", service.ObjectMeta.Namespace))
	req := &ServiceRequest{
		Type:          requestType,
		RemoteService: service,
	}
	s.Events <- req
}

func NewServiceWriter(clientset kubernetes.Interface, events chan *ServiceRequest) *ServiceWriter {
	return &ServiceWriter{
		Events: events,
		Client: clientset,
	}
}

func (s *ServiceWriter) add(svc *v1.Service) {
	logger.Info("Creating service", zap.String("name", svc.Name), zap.String("namespace", svc.ObjectMeta.Namespace))
	s.create(svc)
}

func (s *ServiceWriter) update(svc *v1.Service) {
	logger.Info("Updating service", zap.String("name", svc.Name), zap.String("namespace", svc.ObjectMeta.Namespace))
	update := func() error {
		_, err := s.Client.CoreV1().Services(svc.ObjectMeta.Namespace).Update(svc)
		if err != nil {
			// If the service doesn't exist for some reason, attempt to create it
			if ResourceNotExist(err) {
				s.create(svc)
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

func (s *ServiceWriter) create(svc *v1.Service) {
	create := func() error {
		logger.Info("Creating service", zap.String("name", svc.Name), zap.String("namespace", svc.ObjectMeta.Namespace))
		_, err := s.Client.CoreV1().Services(svc.ObjectMeta.Namespace).Create(svc)
		if err != nil {
			// If the resource already exists, we don't want backoff behavior
			if errors.IsAlreadyExists(err) {
				logger.Info("Service already exists, skipping create",
					zap.String("name", svc.Name),
					zap.String("namespace", svc.ObjectMeta.Namespace))
				return nil
			}
			return err
		}
		return nil
	}
	exponentialBackOff(context.Background(), create)
}

func (s *ServiceWriter) delete(svc *v1.Service) {
	delete := func() error {
		logger.Info("Deleting service", zap.String("name", svc.Name), zap.String("namespace", svc.ObjectMeta.Namespace))
		err := s.Client.CoreV1().Services(svc.ObjectMeta.Namespace).Delete(svc.Name, &metav1.DeleteOptions{})
		if err != nil {
			if PermanentError(err) {
				return backoff.Permanent(err)
			}
			return err
		}
		return nil
	}
	exponentialBackOff(context.Background(), delete)
}

func (s *ServiceWriter) Run() {
	for {
		request := <-s.Events
		switch request.Type {
		case RequestTypeAdd:
			s.add(request.LocalService)
		case RequestTypeUpdate:
			s.update(request.LocalService)
		case RequestTypeDelete:
			s.delete(request.LocalService)
		}
	}
}
