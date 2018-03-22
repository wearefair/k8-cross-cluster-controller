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
//func LocalCoordinator(localClient kubernetes.Interface, localServiceReader chan *k8.ServiceRequest, localEndpointsWriter chan *k8.EndpointsRequest) {
//	for {
//		serviceRequest := <-localServiceReader
//		name := serviceRequest.Service.Name
//		endpoints, err := localClient.CoreV1().
//			Endpoints(serviceRequest.Service.ObjectMeta.Namespace).
//			Get(name, metav1.GetOptions{})
//		if err != nil {
//			errors.Error(context.Background(), err)
//			return
//		}
//		applyEndpointsTransformations(
//			endpoints,
//			applyCrossClusterLabelToEndpoints,
//			sanitizeEndpointsResourceVersion,
//			sanitizeEndpointsUID,
//		)
//		req := &k8.EndpointsRequest{
//			Type:      serviceRequest.Type,
//			Endpoints: endpoints,
//		}
//		localEndpointsWriter <- req
//	}
//}

// Will add in local state if necessary
func ServiceAugmenter(localClient kubernetes.Interface, remoteServiceReader, intermediaryServiceReader chan *k8.ServiceRequest) {
	for {
		req := <-remoteServiceReader
		switch req.Type {
		// Get local state
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

// Just applies the data transformations to the services that are necessary before writing
func ServiceTransformer(intermediaryServiceReader, localServiceWriter chan *k8.ServiceRequest) {
	for {
		req := <-intermediaryServiceReader
	}
}

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

func RemoteCoordinator(remoteServiceReader, localServiceWriter chan *k8.ServiceRequest, remoteEndpointsReader, localEndpointsWriter chan *k8.EndpointsRequest) {
	for {
		select {
		case serviceRequest := <-remoteServiceReader:
			var transformations []serviceTransformer
			// Input sanitization changes depending on Add/Update. Delete requires none
			switch serviceRequest.Type {
			case k8.RequestTypeAdd:
				transformations = []serviceTransformer{
					sanitizeServiceClusterIP,
					sanitizeServiceResourceVersion,
				}
			case k8.RequestTypeUpdate:
				transformations = []serviceTransformer{
					sanitizeServiceClusterIP,
					sanitizeServiceUID,
				}
			}
			applyServiceTransformations(
				serviceRequest.Service,
				transformations...,
			)
			localServiceWriter <- serviceRequest
		case endpointsRequest := <-remoteEndpointsReader:
			var transformations []endpointsTransformer
			switch endpointsRequest.Type {
			case k8.RequestTypeAdd:
				transformations = []endpointsTransformer{
					sanitizeEndpointsResourceVersion,
				}
			case k8.RequestTypeUpdate:
				transformations = []endpointsTransformer{
					sanitizeEndpointsUID,
				}
			}
			applyEndpointsTransformations(
				endpointsRequest.Endpoints,
				transformations...,
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

func sanitizeServiceUID(service *v1.Service) {
	service.ObjectMeta.SetUID("")
}

func sanitizeEndpointsResourceVersion(endpoints *v1.Endpoints) {
	endpoints.ObjectMeta.SetResourceVersion("")
}

func sanitizeEndpointsUID(endpoints *v1.Endpoints) {
	endpoints.ObjectMeta.SetUID("")
}
