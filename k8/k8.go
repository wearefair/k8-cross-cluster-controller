package k8

import (
	"context"
	"time"

	"github.com/wearefair/service-kit-go/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
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

type ServiceRequestType int

type ServiceRequest struct {
	Type     int
	Endpoint *v1.Endpoints
}

// ClusterConfig encapsulates a cluster
type ClusterConfig struct {
	ClusterName string
	Host        string
	Token       string
}

// Client
type Client struct {
	ClusterName string
	K8Client    kubernetes.Interface
	RequestChan chan *ServiceRequest
}

// NewClient returns an instance of Client
func NewClient(clusterConf *ClusterConfig, requestChan chan *ServiceRequest) (*Client, error) {
	var conf *rest.Config
	var err error
	ctx := context.Background()
	if clusterConf == nil {
		// If clusterConf is nil, we're initializing a client within the cluster
		conf, err = rest.InClusterConfig()
	} else {
		conf = &rest.Config{
			Host:        clusterConf.Host,
			BearerToken: clusterConf.Token,
		}
	}
	if err != nil {
		return nil, errors.Error(ctx, err)
	}
	cli, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, errors.Error(ctx, err)
	}
	k8 := &Client{
		ClusterName: clusterConf.ClusterName,
		K8Client:    cli,
		RequestChan: requestChan,
	}
	return k8, nil
}

func (k *Client) WatchAddService(obj interface{}) {
	svc := obj.(*v1.Service)
	endpoint, err := k.getEndpointsFromService(svc)
	if err != nil {
		return
	}
	req := k.createServiceRequest(endpoint, AddService)
	k.RequestChan <- req
}

func (k *Client) WatchUpdateService(old, new interface{}) {
	svc := new.(*v1.Service)
	endpoint, err := k.getEndpointsFromService(svc)
	if err != nil {
		return
	}
	req := k.createServiceRequest(endpoint, UpdateService)
	k.RequestChan <- req
}

func (k *Client) WatchDeleteService(obj interface{}) {
	svc := obj.(*v1.Service)
	endpoint, err := k.getEndpointsFromService(svc)
	if err != nil {
		return
	}
	req := k.createServiceRequest(endpoint, DeleteService)
	k.RequestChan <- req
}

func (k *Client) CreateService(request *ServiceRequest) error {
	return nil
}

func (k *Client) DeleteService(request *ServiceRequest) error {
	return nil
}

func (k *Client) UpdateService(request *ServiceRequest) error {
	return nil
}

func (k *Client) getEndpointsFromService(svc *v1.Service) (*v1.Endpoints, error) {
	ctx := context.Background()
	name := svc.ObjectMeta.Name
	endpoints, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, ferrors.Error(ctx, err)
	}
	return endpoints, nil
}

func (k *Client) createServiceRequest(endpoint *v1.Endpoints, requestType ServiceRequestType) *ServiceRequest {
	return &ServiceRequest{
		ClusterName: k.ClusterName,
		Type:        requestType,
		Endpoint:    endpoint,
	}
}

func WatchServices(k *Client) {
	watchlist := cache.NewListWatchFromClient(k.K8Client, k8Services, defaultNamespace, nil)

	uninitializedWatchList := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.IncludeUninitialized = true
			options.LabelSelector = crossClusterServiceLabel
			return watchlist.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			options.LabelSelector = crossClusterServiceLabel
			return watchlist.Watch(options)
		},
	}

	_, informer := cache.NewInformer(uninitializedWatchList, &v1.Service{}, defaultResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    k.WatchAddService,
			UpdateFunc: k.WatchUpdateService,
			DeleteFunc: k.WatchDeleteService,
		},
	)
}
