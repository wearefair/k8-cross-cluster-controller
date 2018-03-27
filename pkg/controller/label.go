package controller

import (
	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
)

// EndpointsLabel adds a replica label to the local endpoints that's created
func EndpointsLabel(req *k8.EndpointsRequest) error {
	req.LocalEndpoints.ObjectMeta.Labels = labelModifier(req.LocalEndpoints.ObjectMeta.Labels)
	return nil
}

// ServiceLabel adds a replica label to the local service that's created
func ServiceLabel(req *k8.ServiceRequest) error {
	req.LocalService.ObjectMeta.Labels = labelModifier(req.LocalService.ObjectMeta.Labels)
	return nil
}

func labelModifier(labels map[string]string) map[string]string {
	labels[k8.CrossClusterServiceLabelKey] = k8.CrossClusterServiceLocalLabelValue
	return labels
}
