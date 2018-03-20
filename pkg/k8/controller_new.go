package k8

import (
	"k8s.io/api/core/v1"
)

type serviceTransformer func(*v1.Service) *v1.Service

type endpointsTransformer func(*v1.Endpoints) *v1.Endpoints

// We should just be passing in the channels here?
// On the local side, you should have an internalEndpointsWriter and internalEndpointsReader director because all the logic underneath should be at the reader level...
func LocalCoordinator(internalServiceReader chan *ServiceRequest, externalEndpointsWriter chan *EndpointsRequest) {
	for {
		serviceRequest := <-internalServiceReader
	}
}

func RemoteCoordinator(externalServiceReader chan *ServiceRequest, externalEndpointsReader chan *EndpointsRequest, internalServiceWriter chan *ServiceRequest, internalEndpointsWriter chan *EndpointsRequest) {
	for {
	}
}

func applyEndpointsTransformations(transformers ...endpointsTransformer) *v1.Endpoints {
	return nil
}

func applyServiceTransformations(transformers ...serviceTransformer) *v1.Service {
	return nil
}
