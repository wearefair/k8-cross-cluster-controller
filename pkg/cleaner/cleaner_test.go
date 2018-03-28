package cleaner

import (
	"reflect"
	"testing"
	"time"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCleanOrphanedEndpoints(t *testing.T) {
	testCases := []struct {
		LocalEndpoints   []v1.Endpoints
		RemoteEndpoints  []v1.Endpoints
		ExpectedRequests []*k8.EndpointsRequest
	}{
		// If there is 1 local endpoint with 1 remote endpoint that doesn't match, delete local endpoint
		{
			LocalEndpoints: []v1.Endpoints{
				v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			RemoteEndpoints: []v1.Endpoints{
				v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f",
						Namespace: "bar",
					},
				},
			},
			ExpectedRequests: []*k8.EndpointsRequest{
				&k8.EndpointsRequest{
					Type: k8.RequestTypeDelete,
					LocalEndpoints: &v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "bar",
						},
					},
				},
			},
		},
		// If the local endpoint and remote endpoint match, do not delete
		{
			LocalEndpoints: []v1.Endpoints{
				v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			RemoteEndpoints: []v1.Endpoints{
				v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			ExpectedRequests: []*k8.EndpointsRequest{},
		},
		// If there is 1 local endpoint that exist and 2 remote endpoints, 1 that doesn't match, do nothing
		{
			LocalEndpoints: []v1.Endpoints{
				v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			RemoteEndpoints: []v1.Endpoints{
				v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f",
						Namespace: "bar",
					},
				},
				v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			ExpectedRequests: []*k8.EndpointsRequest{},
		},
		// If there is 1 local endpoint and no remote endpoints, delete local endpoint
		{
			LocalEndpoints: []v1.Endpoints{
				v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			RemoteEndpoints: []v1.Endpoints{},
			ExpectedRequests: []*k8.EndpointsRequest{
				&k8.EndpointsRequest{
					Type: k8.RequestTypeDelete,
					LocalEndpoints: &v1.Endpoints{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "bar",
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		cleaner := &Cleaner{
			EndpointWriter: make(chan *k8.EndpointsRequest, 4),
		}
		cleaner.cleanOrphanedEndpoints(testCase.LocalEndpoints, testCase.RemoteEndpoints)
		reqs := []*k8.EndpointsRequest{}
	OUTER:
		for {
			select {
			case <-time.After(time.Millisecond * 30):
				break OUTER
			case req, ok := <-cleaner.EndpointWriter:
				if !ok {
					t.Errorf("Could not read from endpoint writer")
				}
				reqs = append(reqs, req)
			}
		}
		if !reflect.DeepEqual(reqs, testCase.ExpectedRequests) {
			t.Errorf("Expected these endpoints to be queued for deletion %+v\ngot:%+v", testCase.ExpectedRequests, reqs)
		}
	}
}

func TestCleanOrphanedServices(t *testing.T) {
	testCases := []struct {
		LocalService     []v1.Service
		RemoteService    []v1.Service
		ExpectedRequests []*k8.ServiceRequest
	}{
		// If there is 1 local service with 1 remote service that doesn't match, delete local service
		{
			LocalService: []v1.Service{
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			RemoteService: []v1.Service{
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f",
						Namespace: "bar",
					},
				},
			},
			ExpectedRequests: []*k8.ServiceRequest{
				&k8.ServiceRequest{
					Type: k8.RequestTypeDelete,
					LocalService: &v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "bar",
						},
					},
				},
			},
		},
		// If the local service and remote service match, do not delete
		{
			LocalService: []v1.Service{
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			RemoteService: []v1.Service{
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			ExpectedRequests: []*k8.ServiceRequest{},
		},
		// If there is 1 local service that exist and 2 remote service, 1 that doesn't match, do nothing
		{
			LocalService: []v1.Service{
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			RemoteService: []v1.Service{
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "f",
						Namespace: "bar",
					},
				},
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			ExpectedRequests: []*k8.ServiceRequest{},
		},
		// If there is 1 local service and no remote service, delete local service
		{
			LocalService: []v1.Service{
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			RemoteService: []v1.Service{},
			ExpectedRequests: []*k8.ServiceRequest{
				&k8.ServiceRequest{
					Type: k8.RequestTypeDelete,
					LocalService: &v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "bar",
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		cleaner := &Cleaner{
			ServiceWriter: make(chan *k8.ServiceRequest, 4),
		}
		cleaner.cleanOrphanedServices(testCase.LocalService, testCase.RemoteService)
		reqs := []*k8.ServiceRequest{}
	OUTER:
		for {
			select {
			case <-time.After(time.Millisecond * 30):
				break OUTER
			case req, ok := <-cleaner.ServiceWriter:
				if !ok {
					t.Errorf("Could not read from service writer")
				}
				reqs = append(reqs, req)
			}
		}
		if !reflect.DeepEqual(testCase.ExpectedRequests, reqs) {
			t.Errorf("Expected these services to be queued for deletion: %+v, but got %+v", testCase.ExpectedRequests, reqs)
		}
	}
}
