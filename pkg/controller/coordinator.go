package controller

import (
	"context"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type serviceTransformer func(*v1.Service)

type endpointsTransformer func(*v1.Endpoints)

// Local coordinator for picking up local service additions, and tagging endpoints with the appropriate cross cluster label
func LocalCoordinator(client kubernetes.Interface, localServiceReader chan *k8.ServiceRequest, localEndpointsWriter chan *k8.EndpointsRequest) {
	for {
		serviceRequest := <-localServiceReader
		name := serviceRequest.Service.Name
		endpoints, err := client.CoreV1().
			Endpoints(serviceRequest.Service.ObjectMeta.Namespace).
			Get(name, metav1.GetOptions{})
		if err != nil {
			errors.Error(context.Background(), err)
			return
		}
		applyEndpointsTransformations(endpoints, applyCrossClusterLabelToEndpoints)
		req := &k8.EndpointsRequest{
			Type:      serviceRequest.Type,
			Endpoints: endpoints,
		}
		localEndpointsWriter <- req
	}
}

func RemoteCoordinator(remoteServiceReader, localServiceWriter chan *k8.ServiceRequest, remoteEndpointsReader, localEndpointsWriter chan *k8.EndpointsRequest) {
	for {
		select {
		case serviceRequest := <-remoteServiceReader:
			applyServiceTransformations(
				serviceRequest.Service,
				sanitizeServiceClusterIP,
				sanitizeServiceResourceVersion,
			)
			localServiceWriter <- serviceRequest
		case endpointsRequest := <-remoteEndpointsReader:
			applyEndpointsTransformations(
				endpointsRequest.Endpoints,
				sanitizeEndpointsResourceVersion,
				sanitizeEndpointsUID,
			)
			localEndpointsWriter <- endpointsRequest
		}
	}
}

// Chain together multiple functions that alter an endpoints object. Transformations will be applied
// in the order that they're passed in.
func applyEndpointsTransformations(endpoints *v1.Endpoints, transformers ...endpointsTransformer) {
	for _, transformer := range transformers {
		transformer(endpoints)
	}
}

// Chain together multiple functions that alter a service object. Transformations will be applied
// in the order that they're passed in.
func applyServiceTransformations(service *v1.Service, transformers ...serviceTransformer) {
	for _, transformer := range transformers {
		transformer(service)
	}
}

// If the cross cluster label doesn't exist, add it to the endpoints
func applyCrossClusterLabelToEndpoints(endpoints *v1.Endpoints) {
	labels := endpoints.ObjectMeta.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	if _, ok := labels[k8.CrossClusterServiceLabelKey]; !ok {
		labels[k8.CrossClusterServiceLabelKey] = k8.CrossClusterServiceLabelValue
		endpoints.ObjectMeta.SetLabels(labels)
	}
}

func sanitizeServiceResourceVersion(service *v1.Service) {
	service.ObjectMeta.SetResourceVersion("")
}

func sanitizeServiceClusterIP(service *v1.Service) {
	service.Spec.ClusterIP = ""
}

func sanitizeEndpointsResourceVersion(endpoints *v1.Endpoints) {
	endpoints.ObjectMeta.SetResourceVersion("")
}

func sanitizeEndpointsUID(endpoints *v1.Endpoints) {
	endpoints.ObjectMeta.SetUID("")
}
