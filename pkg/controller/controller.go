package controller

import (
	"context"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAugmenter takes a K8 client and two service channels. On a service update from the remote side, it'll augment the request by
// adding in a local version of the service and then pass it down to the transformer channel
func ServiceAugmenter(localClient kubernetes.Interface, remoteServiceReader, intermediaryServiceReader chan *k8.ServiceRequest) {
	for {
		req := <-remoteServiceReader
		switch req.Type {
		case k8.RequestTypeAdd:
			req.LocalService = &v1.Service{}
		case k8.RequestTypeUpdate:
			localService, err := localClient.CoreV1().Services(req.RemoteService.ObjectMeta.Namespace).Get(req.RemoteService.Name, metav1.GetOptions{})
			if err != nil {
				errors.Error(context.Background(), err)
				continue
			}
			req.LocalService = localService
		}
		intermediaryServiceReader <- req
	}
}

// ServiceTransformer reads from a service channel and filters service metadata and spec that we don't
// want propagated from the remote cluster to the local cluster, and passes it onto the local writer for
// resource creation
func ServiceTransformer(intermediaryServiceReader, localServiceWriter chan *k8.ServiceRequest) {
	for {
		req := <-intermediaryServiceReader
		serviceWhitelist(req.RemoteService, req.LocalService)
		localServiceWriter <- req
	}
}

// serviceWhitelist allows only the fields that we want to allow to be copied over. Metadata such as UID and
// resourceVersion cannot be propagated from one K8 cluster to another on creates/updates
func serviceWhitelist(oldSvc, newSvc *v1.Service) {
	newSvc.ObjectMeta = objectMetaWhitelist(oldSvc.ObjectMeta, newSvc.ObjectMeta)
	newSvc.Spec.Ports = oldSvc.Spec.Ports
	newSvc.Spec.SessionAffinity = oldSvc.Spec.SessionAffinity
}

// EndpointsAugmenter takes a K8 client and two endpoints channels. On an endpoint update from the remote side, it'll augment the request by
// adding in a local version of the endpoint and then pass it down to the transformer channel
func EndpointsAugmenter(localClient kubernetes.Interface, remoteEndpointsReader, intermediaryEndpointsReader chan *k8.EndpointsRequest) {
	for {
		req := <-remoteEndpointsReader
		switch req.Type {
		case k8.RequestTypeAdd:
			req.LocalEndpoints = &v1.Endpoints{}
		case k8.RequestTypeUpdate:
			localEndpoints, err := localClient.CoreV1().Endpoints(req.RemoteEndpoints.ObjectMeta.Namespace).Get(req.RemoteEndpoints.Name, metav1.GetOptions{})
			if err != nil {
				errors.Error(context.Background(), err)
				continue
			}
			req.LocalEndpoints = localEndpoints
		}
		intermediaryEndpointsReader <- req
	}
}

// EndpointsTransformer reads from an endpoints channel and filters endpoints metadata and spec that we don't
// want propagated from the remote cluster to the local cluster, and passes it onto the local writer for
// resource creation
func EndpointsTransformer(intermediaryEndpointsReader, localEndpointsWriter chan *k8.EndpointsRequest) {
	for {
		req := <-intermediaryEndpointsReader
		endpointsWhitelist(req.RemoteEndpoints, req.LocalEndpoints)
		localEndpointsWriter <- req
	}
}

// endpointsWhitelist allows only the fields that we want to be copied over. Metadata such as UID and
// resourceVersion cannot be propagated from one K8 cluster to another on creates/updates
func endpointsWhitelist(oldEndpoints, newEndpoints *v1.Endpoints) {
	newEndpoints.ObjectMeta = objectMetaWhitelist(oldEndpoints.ObjectMeta, newEndpoints.ObjectMeta)
	newEndpoints.Subsets = oldEndpoints.Subsets
}

func objectMetaWhitelist(oldMeta, newMeta metav1.ObjectMeta) metav1.ObjectMeta {
	newMeta.Name = oldMeta.Name
	newMeta.Namespace = oldMeta.Namespace
	newMeta.Labels = oldMeta.Labels
	return newMeta
}
