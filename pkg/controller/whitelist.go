package controller

import (
	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceWhitelist allows only the fields that we want to allow to be copied over. Metadata such as UID and
// resourceVersion cannot be propagated from one K8 cluster to another on creates/updates
func ServiceWhitelist(req *k8.ServiceRequest) error {
	req.LocalService.ObjectMeta = objectMetaWhitelist(req.RemoteService.ObjectMeta, req.LocalService.ObjectMeta)
	req.LocalService.Spec.Ports = req.RemoteService.Spec.Ports
	req.LocalService.Spec.SessionAffinity = req.RemoteService.Spec.SessionAffinity
	return nil
}

// EndpointsWhitelist allows only the fields that we want to be copied over. Metadata such as UID and
// resourceVersion cannot be propagated from one K8 cluster to another on creates/updates
func EndpointsWhitelist(req *k8.EndpointsRequest) error {
	req.LocalEndpoints.ObjectMeta = objectMetaWhitelist(req.RemoteEndpoints.ObjectMeta, req.LocalEndpoints.ObjectMeta)
	req.LocalEndpoints.Subsets = req.RemoteEndpoints.Subsets
	return nil
}

func objectMetaWhitelist(remoteMeta, localMeta metav1.ObjectMeta) metav1.ObjectMeta {
	localMeta.Name = remoteMeta.Name
	localMeta.Namespace = remoteMeta.Namespace
	localMeta.Labels = remoteMeta.Labels
	return localMeta
}
