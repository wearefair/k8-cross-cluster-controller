package k8

import (
	"fmt"

	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/api/core/v1"
)

const (
	RequestTypeAdd RequestType = iota
	RequestTypeUpdate
	RequestTypeDelete

	CrossClusterServiceLabelKey   = "fair.com/cross-cluster"
	CrossClusterServiceLabelValue = "true"
)

var (
	logger         = logging.Logger()
	RequestTypeMap = map[RequestType]string{
		RequestTypeAdd:    "add",
		RequestTypeUpdate: "update",
		RequestTypeDelete: "delete",
	}
	CrossClusterLabel = fmt.Sprintf("%s=%s", CrossClusterServiceLabelKey, CrossClusterServiceLabelValue)
)

type RequestType int

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
