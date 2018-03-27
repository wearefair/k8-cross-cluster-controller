package k8

import (
	"fmt"

	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RequestType int

const (
	RequestTypeAdd RequestType = iota
	RequestTypeUpdate
	RequestTypeDelete

	CrossClusterServiceLabelKey         = "fair.com/cross-cluster"
	CrossClusterServiceLocalLabelValue  = "replicated"
	CrossClusterServiceRemoteLabelValue = "replicate"
)

var (
	logger = logging.Logger()

	CrossClusterLocalLabel  = fmt.Sprintf("%s=%s", CrossClusterServiceLabelKey, CrossClusterServiceLocalLabelValue)
	CrossClusterRemoteLabel = fmt.Sprintf("%s=%s", CrossClusterServiceLabelKey, CrossClusterServiceRemoteLabelValue)
	LocalFilter             = metav1.ListOptions{
		LabelSelector: CrossClusterLocalLabel,
	}
	RemoteFilter = func(options *metav1.ListOptions) {
		options.LabelSelector = CrossClusterRemoteLabel
	}
	RequestTypeMap = map[RequestType]string{
		RequestTypeAdd:    "add",
		RequestTypeUpdate: "update",
		RequestTypeDelete: "delete",
	}
)

type ServiceRequest struct {
	Type          RequestType
	RemoteService *v1.Service
	LocalService  *v1.Service
}

type EndpointsRequest struct {
	Type            RequestType
	RemoteEndpoints *v1.Endpoints
	LocalEndpoints  *v1.Endpoints
}
