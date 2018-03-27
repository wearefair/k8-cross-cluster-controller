package controller

import "github.com/wearefair/k8-cross-cluster-controller/pkg/k8"

func EndpointsPipeline(in, out chan *k8.EndpointsRequest, transformers ...EndpointsTransformer) {
OUTER:
	for {
		req := <-in
		for _, transformer := range transformers {
			// We've already reported the error. However, we don't want anything to fail here, so we
			// just skip trying to create the request
			if err := transformer(req); err != nil {
				continue OUTER
			}
		}
		out <- req
	}
}

func ServicePipeline(in, out chan *k8.ServiceRequest, transformers ...ServiceTransformer) {
OUTER:
	for {
		req := <-in
		for _, transformer := range transformers {
			// We've already reported the error. However, we don't want anything to fail here, so we
			// just skip trying to create the request
			if err := transformer(req); err != nil {
				continue OUTER
			}
		}
		out <- req
	}
}
