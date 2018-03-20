package k8

import (
	"context"

	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Converter receives and converts services to endpoints
type Converter struct {
	// Reader
	ServiceEvents chan *ServicesRequest
	// Funnel to another reader chan maybe
	EndpointsEvents chan *EndpointsRequest
	Client          kubernetes.Interface
}

func NewConverter(clientset kubernetes.Interface) *Convert {
	return &Converter{
		Events: make(chan *ServicesRequest),
	}
}

func (c *Converter) Run() {
}

func (c *Converter) getEndpointsFromService(svc *v1.Service) (*v1.Endpoints, error) {
	name := svc.ObjectMeta.Name
	endpoints, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Error(context.Background(), err)
	}
	return endpoints, nil
}
