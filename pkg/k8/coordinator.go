package k8

import (
	"context"

	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type serviceTransformer func(*v1.Service) *v1.Service

type endpointsTransformer func(*v1.Endpoints) *v1.Endpoints

// Local coordinator for picking up local service additions, and tagging endpoints with the appropriate cross cluster label
func LocalCoordinator(client kubernetes.Interface, localServiceReader chan *ServiceRequest, localEndpointsWriter chan *EndpointsRequest) {
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
		req := &EndpointsRequest{
			Type:      serviceRequest.Type,
			Endpoints: endpoints,
		}
		localEndpointsWriter <- req
	}
}

func RemoteCoordinator(remoteServiceReader, localServiceWriter chan *ServiceRequest, remoteEndpointsReader, localEndpointsWriter chan *EndpointsRequest) {
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
				sanitizeEndpointResourceVersion,
				sanitizeEndpointsUID,
			)
			localEndpointsWriter <- endpointsRequest
		}
	}
}

func applyEndpointsTransformations(endpoints *v1.Endpoints, transformers ...endpointsTransformer) *v1.Endpoints {
	for _, transformer := range transformers {
		endpoints = transformer(endpoints)
	}
	return endpoints
}

func applyServiceTransformations(service *v1.Service, transformers ...serviceTransformer) *v1.Service {
	for _, transformer := range transformers {
		service = transformer(service)
	}
	return service
}

// I don't think I need to pass back an endpoint because it's applying the transformation...
func applyCrossClusterLabelToEndpoints(endpoints *v1.Endpoints) *v1.Endpoints {
	labels := endpoints.ObjectMeta.GetLabels()
	// If the cross cluster label doesn't exist, add it to the endpoints
	if _, ok := labels[CrossClusterServiceLabelKey]; !ok {
		labels[CrossClusterServiceLabelKey] = CrossClusterServiceLabelValue
		endpoints.ObjectMeta.SetLabels(labels)
	}
	return endpoints
}

func sanitizeServiceResourceVersion(service *v1.Service) *v1.Service {
	service.ObjectMeta.SetResourceVersion("")
	return service
}

func sanitizeServiceClusterIP(service *v1.Service) *v1.Service {
	service.Spec.ClusterIP = ""
	return service
}

func sanitizeEndpointResourceVersion(endpoints *v1.Endpoints) *v1.Endpoints {
	endpoints.ObjectMeta.SetResourceVersion("")
	return endpoints
}

func sanitizeEndpointsUID(endpoints *v1.Endpoints) *v1.Endpoints {
	endpoints.ObjectMeta.SetUID("")
	return endpoints
}
