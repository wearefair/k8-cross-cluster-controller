package k8

import (
	"context"

	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterConfig encapsulates a cluster
// TODO turn this into a file
type ClusterConfig struct {
	Cluster string `yaml:"cluster"`
	Host    string `yaml:"host"`
	Token   string `yaml:"token"`
}

// TODO: Split this out into ingress cluster and maybe egress cluster?
// Or cross-cluster client and internal client
type InternalClient struct {
	Cluster     string
	K8Client    kubernetes.Interface
	RequestChan chan *ServiceRequest
}

// NewClient returns an instance of Client
func NewInternalClient(clientset kubernetes.Interface, requestChan chan *ServiceRequest) *InternalClient {
	return &InternalClient{
		K8Client:    clientset,
		RequestChan: requestChan,
	}
}

// CreateService creates an endpoints object and a corresponding service on the cross cluster
func (k *InternalClient) CreateService(request *ServiceRequest) error {
	endpoints, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Create(request.Endpoint)
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	if _, err := k.createServiceFromEndpoints(endpoints); err != nil {
		return err
	}
	return nil
}

// DeleteService takes a ServiceRequest and deletes both the service and the endpoint object
func (k *InternalClient) DeleteService(request *ServiceRequest) error {
	ctx := context.Background()
	err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Delete(request.Endpoint.ObjectMeta.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Error(ctx, err)
	}
	err = k.K8Client.CoreV1().Services(defaultNamespace).Delete(request.Endpoint.ObjectMeta.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Error(ctx, err)
	}
	return nil
}

func (k *InternalClient) UpdateService(request *ServiceRequest) error {
	_, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Update(request.Endpoint)
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	return nil
}

func (k *InternalClient) createServiceFromEndpoints(endpoints *v1.Endpoints) (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: endpoints.ObjectMeta.Name,
		},
	}
	svc, err := k.K8Client.CoreV1().Services(defaultNamespace).Create(svc)
	if err != nil {
		return nil, errors.Error(context.Background(), err)
	}
	return svc, nil
}
