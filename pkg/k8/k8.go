package k8

import (
	"time"

	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/api/core/v1"
)

const (
	RequestTypeAdd RequestType = iota
	RequestTypeUpdate
	RequestTypeDelete

	CrossClusterServiceLabelKey   = "fair.com/cross-cluster"
	CrossClusterServiceLabelValue = "true"
	defaultNamespace              = "default"
	defaultResyncPeriod           = 30 * time.Second
	k8Endpoints                   = "endpoints"
	k8Services                    = "services"
)

var (
	logger         = logging.Logger()
	RequestTypeMap = map[RequestType]string{
		RequestTypeAdd:    "add",
		RequestTypeUpdate: "update",
		RequestTypeDelete: "delete",
	}
)

type RequestType int

type ServiceRequest struct {
	Type         RequestType
	Service      *v1.Service
	LocalService *v1.Service
}

type EndpointsRequest struct {
	Type           RequestType
	Endpoints      *v1.Endpoints
	LocalEndpoints *v1.Endpoints
}
