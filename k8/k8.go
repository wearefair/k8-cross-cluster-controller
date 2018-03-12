package k8

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	Type     ServiceRequestType
	Endpoint *v1.Endpoints
}

// ClusterConfig encapsulates a cluster
// TODO turn this into a file
type ClusterConfig struct {
	ClusterName string
	Host        string
	Token       string
}

// TODO: Split this out into ingress cluster and maybe egress cluster?
// Or cross-cluster client and internal client
type Client struct {
	K8Config    *rest.Config
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
		K8Config:    conf,
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

// CreateService creates an endpoints object and a corresponding service on the cross cluster
func (k *Client) CreateService(request *ServiceRequest) error {
	endpoints, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Create(request.Endpoint)
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	if _, err := k.createServiceFromEndpoints(endpoints); err != nil {
		return err
	}
	return nil
}

// DeleteService takes a ServiceRequest and deletes both the service and the endpoint object
func (k *Client) DeleteService(request *ServiceRequest) error {
	ctx := context.Background()
	err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Delete(request.Endpoint.ObjectMeta.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Error(ctx, err)
	}
	err = k.K8Client.CoreV1().Services(defaultNamespace).Delete(request.Endpoint.ObjectMeta.Name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Error(ctx, err)
	}
	return nil
}

func (k *Client) UpdateService(request *ServiceRequest) error {
	ctx := context.Background()
	_, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Update(request.Endpoint)
	if err != nil {
		return errors.Error(ctx, err)
	}
	// I don't think the service needs to be updated if the endpoints already are?
	return nil
}

func (k *Client) getEndpointsFromService(svc *v1.Service) (*v1.Endpoints, error) {
	name := svc.ObjectMeta.Name
	endpoints, err := k.K8Client.CoreV1().Endpoints(defaultNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Error(context.Background(), err)
	}
	return endpoints, nil
}

func (k *Client) createServiceFromEndpoints(endpoints *v1.Endpoints) (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: endpoints.ObjectMeta.Name,
		},
	}
	svc, err := k.K8Client.CoreV1().Services(defaultNamespace).Create(svc)
	if err != nil {
		return nil, errors.Error(context.Background(), err)
	}
	return svc, nil
}

func (k *Client) createServiceRequest(endpoint *v1.Endpoints, requestType ServiceRequestType) *ServiceRequest {
	return &ServiceRequest{
		Type:     requestType,
		Endpoint: endpoint,
	}
}

func WatchServices(k *Client) error {
	// This needs to be a RESTClient
	restClient, err := rest.RESTClientFor(k.K8Config)
	if err != nil {
		return errors.Error(context.Background(), err)
	}
	watchlist := cache.NewListWatchFromClient(restClient, k8Services, defaultNamespace, nil)

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
	stop := make(chan struct{})
	go informer.Run(stop)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	fmt.Println("Shutting down")
	close(stop)
	return nil
}
