package controller

import (
	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EndpointsLabel adds a replica label to the local endpoints that's created
func EndpointsLabel(req *k8.EndpointsRequest) error {
	labelModifier(req.LocalEndpoints.ObjectMeta)
	return nil
}

// ServiceLabel adds a replica label to the local service that's created
func ServiceLabel(req *k8.ServiceRequest) error {
	labelModifier(req.LocalService.ObjectMeta)
	return nil
}

func labelModifier(meta metav1.ObjectMeta) {
	meta.Labels[k8.CrossClusterServiceLabelKey] = k8.CrossClusterServiceLocalLabelValue
}
