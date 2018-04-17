package k8

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	ferrors "github.com/wearefair/service-kit-go/errors"
	"github.com/wearefair/service-kit-go/logging"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RequestType int

const (
	backOffMaxElapsedTime             = 2 * time.Minute
	backOffMaxInterval                = 5 * time.Second
	RequestTypeAdd        RequestType = iota
	RequestTypeUpdate
	RequestTypeDelete

	CrossClusterServiceLabelKey         = "fair.com/cross-cluster"
	CrossClusterServiceLocalLabelValue  = "follower"
	CrossClusterServiceRemoteLabelValue = "true"
)

var (
	logger = logging.Logger()

	CrossClusterLocalLabel  = fmt.Sprintf("%s=%s", CrossClusterServiceLabelKey, CrossClusterServiceLocalLabelValue)
	CrossClusterRemoteLabel = fmt.Sprintf("%s=%s", CrossClusterServiceLabelKey, CrossClusterServiceRemoteLabelValue)
	// Ideally the local filter and the remote filter will be in a consistent format, but the watcher for the remote filter
	// requires an options func
	LocalFilter = metav1.ListOptions{
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

func ResourceNotExist(err error) bool {
	return errors.IsNotFound(err) || errors.IsGone(err)
}

func exponentialBackOff(ctx context.Context, retryFunc func() error) {
	// Get settings and then override the ones we don't want
	settings := backoff.NewExponentialBackOff()
	settings.MaxInterval = backOffMaxInterval
	settings.MaxElapsedTime = backOffMaxElapsedTime
	err := backoff.Retry(retryFunc, settings)
	if err != nil {
		ferrors.Error(ctx, err)
	}
}
