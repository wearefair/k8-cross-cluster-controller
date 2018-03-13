package k8

import (
	"time"

	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/api/core/v1"
)

const (
	AddService ServiceRequestType = iota
	UpdateService
	DeleteService

	crossClusterServiceLabel = "fair.com/cross-cluster=true"
	defaultNamespace         = "default"
	defaultResyncPeriod      = 30 * time.Second
	k8Services               = "services"
)

var (
	logger = logging.Logger()
)

type ServiceRequestType int

type ServiceRequest struct {
	Type     ServiceRequestType
	Endpoint *v1.Endpoints
}
