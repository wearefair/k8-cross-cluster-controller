package cleaner

import (
	"context"
	"time"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	ferrors "github.com/wearefair/service-kit-go/errors"
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

// Cleaner looks for orphaned services/endpoints on the local side and sends requests to delete them.
// A service/endpoint is considered an orphan if it's no longer existing on the remote side, but somehow still exists
// on the local side
type Cleaner struct {
	LocalClient    kubernetes.Interface
	RemoteClient   kubernetes.Interface
	EndpointWriter chan *k8.EndpointsRequest
	ServiceWriter  chan *k8.ServiceRequest
}

func New(localClient, remoteClient kubernetes.Interface, endpointWriter chan *k8.EndpointsRequest, serviceWriter chan *k8.ServiceRequest) *Cleaner {
	return &Cleaner{
		LocalClient:    localClient,
		RemoteClient:   remoteClient,
		EndpointWriter: endpointWriter,
		ServiceWriter:  serviceWriter,
	}
}

func (c *Cleaner) Run(stopChan <-chan struct{}) {
	logger.Info("Starting cleaner")
	ticker := time.NewTicker(defaultSleepTime)
	defer ticker.Stop()
	for {
		select {
		case <-stopChan:
			logger.Info("Received stopped signal. Stopping clean")
			return
		case <-ticker.C:
			// If there is a service or endpoint that's local that no longer exists on remote side, send a deletion event
			c.cleanOrphanedServices(c.listLocalServices(), c.listRemoteServices())
			c.cleanOrphanedEndpoints(c.listLocalEndpoints(), c.listRemoteEndpoints())
		}
	}
}

func (c *Cleaner) cleanOrphanedServices(localServices, remoteServices []v1.Service) {
	for _, localService := range localServices {
		if exists := c.checkServiceExists(localService, remoteServices); !exists {
			req := &k8.ServiceRequest{
				Type:         k8.RequestTypeDelete,
				LocalService: &localService,
			}
			c.ServiceWriter <- req
		}
	}
}

func (c *Cleaner) checkServiceExists(localService v1.Service, remoteServices []v1.Service) bool {
	for _, remoteService := range remoteServices {
		if (remoteService.ObjectMeta.Namespace == localService.ObjectMeta.Namespace) && (localService.Name == remoteService.Name) {
			return true
		}
	}
	return false
}

func (c *Cleaner) cleanOrphanedEndpoints(localEndpoints, remoteEndpoints []v1.Endpoints) {
	for _, localEndpoint := range localEndpoints {
		if exists := c.checkEndpointsExists(localEndpoint, remoteEndpoints); !exists {
			req := &k8.EndpointsRequest{
				Type:           k8.RequestTypeDelete,
				LocalEndpoints: &localEndpoint,
			}
			c.EndpointWriter <- req
		}
	}
}

func (c *Cleaner) checkEndpointsExists(localEndpoint v1.Endpoints, remoteEndpoints []v1.Endpoints) bool {
	for _, remoteEndpoint := range remoteEndpoints {
		if (remoteEndpoint.ObjectMeta.Namespace == localEndpoint.ObjectMeta.Namespace) && (localEndpoint.Name == remoteEndpoint.Name) {
			return true
		}
	}
	return false
}

// Lists all endpoints that are local with the cross cluster label
func (c *Cleaner) listLocalEndpoints() []v1.Endpoints {
	logger.Info("Listing local endpoints for clean")
	list, err := c.LocalClient.CoreV1().Endpoints(metav1.NamespaceAll).List(k8.LocalFilter)
	// If there's an error, we want to report it, but we don't necessarily need to propagate it
	if err != nil {
		ferrors.Error(context.Background(), err)
		return []v1.Endpoints{}
	}
	return list.Items
}

// List all services that are local with the cross cluster label
func (c *Cleaner) listLocalServices() []v1.Service {
	logger.Info("Listing local services for clean")
	list, err := c.LocalClient.CoreV1().Services(metav1.NamespaceAll).List(k8.LocalFilter)
	// If there's an error, we want to report it, but we don't necessarily need to propagate it
	if err != nil {
		ferrors.Error(context.Background(), err)
		return []v1.Service{}
	}
	return list.Items
}

// Lists all services that are remote with the cross cluster label
func (c *Cleaner) listRemoteServices() []v1.Service {
	opts := &metav1.ListOptions{}
	k8.RemoteFilter(opts)
	logger.Info("Listing remote services for clean")
	list, err := c.RemoteClient.CoreV1().Services(metav1.NamespaceAll).List(*opts)
	// If there's an error, we want to report it, but we don't necessarily need to propagate it
	if err != nil {
		ferrors.Error(context.Background(), err)
		return []v1.Service{}
	}
	return list.Items
}

// Lists all endpoints that are remote with the cross cluster label
func (c *Cleaner) listRemoteEndpoints() []v1.Endpoints {
	opts := &metav1.ListOptions{}
	k8.RemoteFilter(opts)
	logger.Info("Listing remote services for clean")
	list, err := c.RemoteClient.CoreV1().Endpoints(metav1.NamespaceAll).List(*opts)
	// If there's an error, we want to report it, but we don't necessarily need to propagate it
	if err != nil {
		ferrors.Error(context.Background(), err)
		return []v1.Endpoints{}
	}
	return list.Items
}
