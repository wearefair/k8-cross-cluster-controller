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
		case k8.RequestTypeUpdate:
			localService, err := localClient.CoreV1().Services(req.Service.ObjectMeta.Namespace).Get(req.Service.Name, metav1.GetOptions{})
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
		switch req.Type {
		case k8.RequestTypeAdd:
			req.Service = serviceWhitelist(req.Service, &v1.Service{})
		case k8.RequestTypeUpdate:
			req.Service = serviceWhitelist(req.Service, req.LocalService)
		}
		localServiceWriter <- req
	}
}

// serviceWhitelist allows only the fields that we want to allow to be copied over. Metadata such as UID and
// resourceVersion cannot be propagated from one K8 cluster to another on creates/updates
func serviceWhitelist(oldSvc, newSvc *v1.Service) *v1.Service {
	newSvc.ObjectMeta.Name = oldSvc.ObjectMeta.Name
	newSvc.ObjectMeta.Namespace = oldSvc.ObjectMeta.Namespace
	newSvc.ObjectMeta.Labels = oldSvc.ObjectMeta.Labels
	newSvc.ObjectMeta.ClusterName = oldSvc.ObjectMeta.ClusterName
	newSvc.Spec.Ports = oldSvc.Spec.Ports
	newSvc.Spec.SessionAffinity = oldSvc.Spec.SessionAffinity
	return newSvc
}

// EndpointsAugmenter takes a K8 client and two endpoints channels. On an endpoint update from the remote side, it'll augment the request by
// adding in a local version of the endpoint and then pass it down to the transformer channel
func EndpointsAugmenter(localClient kubernetes.Interface, remoteEndpointsReader, intermediaryEndpointsReader chan *k8.EndpointsRequest) {
	for {
		req := <-remoteEndpointsReader
		switch req.Type {
		case k8.RequestTypeUpdate:
			localEndpoints, err := localClient.CoreV1().Endpoints(req.Endpoints.ObjectMeta.Namespace).Get(req.Endpoints.Name, metav1.GetOptions{})
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
		switch req.Type {
		case k8.RequestTypeAdd:
			req.Endpoints = endpointsWhitelist(req.Endpoints, &v1.Endpoints{})
		case k8.RequestTypeUpdate:
			req.Endpoints = endpointsWhitelist(req.Endpoints, req.LocalEndpoints)
		}
		localEndpointsWriter <- req
	}
}

// endpointsWhitelist allows only the fields that we want to be copied over. Metadata such as UID and
// resourceVersion cannot be propagated from one K8 cluster to another on creates/updates
func endpointsWhitelist(oldEndpoints, newEndpoints *v1.Endpoints) *v1.Endpoints {
	newEndpoints.ObjectMeta.Name = oldEndpoints.ObjectMeta.Name
	newEndpoints.ObjectMeta.Namespace = oldEndpoints.ObjectMeta.Namespace
	newEndpoints.ObjectMeta.Labels = oldEndpoints.ObjectMeta.Labels
	newEndpoints.ObjectMeta.ClusterName = oldEndpoints.ObjectMeta.ClusterName
	newEndpoints.Subsets = oldEndpoints.Subsets
	return newEndpoints
}
