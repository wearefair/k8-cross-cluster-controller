package controller

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	"github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	defaultSleepTime = 5 * time.Minute
)

var (
	logger = logging.Logger()
)

// Cleaner just looks for any orphaned local services and endpoints and sends requests to
// the local writers to clean them up
type Cleaner struct {
	LocalClient    kubernetes.Interface
	RemoteClient   kubernetes.Interface
	EndpointWriter chan *k8.EndpointsRequest
	ServiceWriter  chan *k8.ServiceRequest
}

func NewCleaner(localClient, remoteClient kubernetes.Interface, endpointWriter chan *k8.EndpointsRequest, serviceWriter chan *k8.ServiceRequest) *Cleaner {
	return &Cleaner{
		LocalClient:    localClient,
		RemoteClient:   remoteClient,
		EndpointWriter: endpointWriter,
		ServiceWriter:  serviceWriter,
	}
}

// TODO: Stop when message is sent over stop channel
func (c *Cleaner) Run(stopChan <-chan struct{}) {
	for {
		// If there is a service that's local that no longer exists on remote side, delete it
		services := c.listLocalServices()
		if services != nil {
			for _, service := range services {
				_, err := c.RemoteClient.CoreV1().Services(service.ObjectMeta.Namespace).Get(service.Name, metav1.GetOptions{})
				// If err != nil, likely svc doesn't exist, so delete it?
				if err != nil {
					logger.Warn("Error retrieving remote service, likely it doesn't exist, deleting...", zap.Error(err), zap.String("service", service.Name))
					// Send down channel to delete
					req := &k8.ServiceRequest{
						Type:    k8.RequestTypeDelete,
						Service: &service,
					}
					c.ServiceWriter <- req
				}
			}
		}
		// If there is an endpoint that exists on local side, but no longer on remote side, delete it
		endpoints := c.listLocalEndpoints()
		if endpoints != nil {
			for _, endpoint := range endpoints {
				_, err := c.RemoteClient.CoreV1().Endpoints(endpoint.ObjectMeta.Namespace).Get(endpoint.Name, metav1.GetOptions{})
				if err != nil {
					logger.Warn("Error retrieving remote endpoint, likely it doesn't exist, deleting...", zap.Error(err), zap.String("endpoint", endpoint.Name))
					// Send down channel to delete
					req := &k8.EndpointsRequest{
						Type:      k8.RequestTypeDelete,
						Endpoints: &endpoint,
					}
					c.EndpointWriter <- req
				}
			}
		}
	}
}

// Lists all endpoints that are local with the cross cluster label
func (c *Cleaner) listLocalEndpoints() []v1.Endpoints {
	opts := metav1.ListOptions{
		LabelSelector: k8.CrossClusterLabel,
	}
	list, err := c.LocalClient.CoreV1().Endpoints(metav1.NamespaceAll).List(opts)
	if err != nil {
		errors.Error(context.Background(), err)
		return nil
	}
	return list.Items
}

// List all services that are local with the cross cluster label
func (c *Cleaner) listLocalServices() []v1.Service {
	opts := metav1.ListOptions{
		LabelSelector: k8.CrossClusterLabel,
	}
	list, err := c.LocalClient.CoreV1().Services(metav1.NamespaceAll).List(opts)
	// If there's an error, we want to report it, but we don't necessarily need to propagate it
	if err != nil {
		errors.Error(context.Background(), err)
		return nil
	}
	return list.Items
}
