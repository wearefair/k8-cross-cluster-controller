package controller

import (
	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
)

type EndpointsTransformer func(req *k8.EndpointsRequest) error

type ServiceTransformer func(req *k8.ServiceRequest) error
