package controller

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/apimachinery/pkg/api/errors"

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

// Cleaner looks for orphaned services/endpoints on the local side and sends requests to delete them.
// A service/endpoint is considered an orphan if it's no longer existing on the remote side, but somehow still exists
// on the local side
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

func (c *Cleaner) Run(stopChan <-chan struct{}) {
	ticker := time.NewTicker(defaultSleepTime).C
	for {
		select {
		case <-stopChan:
			logger.Info("Received stopped signal. Stopping clean")
			return
		case <-ticker:
			// If there is a service that's local that no longer exists on remote side, send a deletion event
			services := c.listLocalServices()
			for _, service := range services {
				_, err := c.RemoteClient.CoreV1().Services(service.ObjectMeta.Namespace).Get(service.Name, metav1.GetOptions{})
				if err != nil {
					if errors.IsGone(err) || errors.NotFound(err) {
						logger.Info("Deleting remote service", zap.Error(err), zap.String("service", service.Name))
						// Send down channel to delete
						req := &k8.ServiceRequest{
							Type:    k8.RequestTypeDelete,
							Service: &service,
						}
						c.ServiceWriter <- req
					} else {
						ferrors.Error(context.Background(), err)
					}
				}
			}
			// If there is an endpoint that exists on local side, but no longer on remote side, send a deletion event
			endpoints := c.listLocalEndpoints()
			for _, endpoint := range endpoints {
				_, err := c.RemoteClient.CoreV1().Endpoints(endpoint.ObjectMeta.Namespace).Get(endpoint.Name, metav1.GetOptions{})
				if err != nil {
					if errors.IsGone(err) || errors.NotFound(err) {
						logger.Info("Deleting remote endpoint", zap.Error(err), zap.String("endpoint", endpoint.Name))
						// Send down channel to delete
						req := &k8.EndpointsRequest{
							Type:      k8.RequestTypeDelete,
							Endpoints: &endpoint,
						}
						c.EndpointWriter <- req
					} else {
						ferrors.Error(context.Background(), err)
					}
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
		ferrors.Error(context.Background(), err)
		return []v1.Endpoints{}
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
		ferrors.Error(context.Background(), err)
		return []v1.Service{}
	}
	return list.Items
}
