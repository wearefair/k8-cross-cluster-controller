package k8

import "k8s.io/api/core/v1"

// TODO: Ungarbage these and make these just copy over the things we want
// such as name and labels

// SanitizeService takes a service object and sets ClusterIP to empty
// and removes resourceVersion.
func SanitizeService(svc *v1.Service) *v1.Service {
	svc.ObjectMeta.SetResourceVersion("")
	svc.Spec.ClusterIP = ""
	return svc
}

// SanitizeEndpoints takes an endpoint object and removes
// resourceVersion and UID
func SanitizeEndpoints(endpoints *v1.Endpoints) *v1.Endpoints {
	endpoints.ObjectMeta.SetResourceVersion("")
	endpoints.ObjectMeta.SetUID("")
	return endpoints
}
