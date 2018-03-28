package controller

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Conflating the Endpoints and Service function tests together since they're virtually identical
func TestEndpointsServiceLabel(t *testing.T) {
	testCases := []struct {
		Labels   map[string]string
		Expected map[string]string
	}{
		// Empty labels gets the cross cluster key and "replicated" value
		{
			Labels: map[string]string{},
			Expected: map[string]string{
				k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLocalLabelValue,
			},
		},
		// Label with the cross cluster key but "replicate" value gets "replicated value"
		{
			Labels: map[string]string{
				k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceRemoteLabelValue,
			},
			Expected: map[string]string{
				k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLocalLabelValue,
			},
		},
		// Label with random other key and values gets the cross cluster key and "replicated" value
		{
			Labels: map[string]string{
				"foo": "bar",
			},
			Expected: map[string]string{
				"foo": "bar",
				k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLocalLabelValue,
			},
		},
	}

	for _, testCase := range testCases {
		endpointReq := &k8.EndpointsRequest{
			LocalEndpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: testCase.Labels,
				},
			},
		}
		serviceReq := &k8.ServiceRequest{
			LocalService: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: testCase.Labels,
				},
			},
		}
		EndpointsLabel(endpointReq)
		ServiceLabel(serviceReq)
		if !reflect.DeepEqual(endpointReq.LocalEndpoints.ObjectMeta.Labels, testCase.Expected) {
			t.Errorf("Expected endpoints labels: %v\ngot: %v", testCase.Expected, endpointReq.LocalEndpoints.ObjectMeta.Labels)
		}
		if !reflect.DeepEqual(serviceReq.LocalService.ObjectMeta.Labels, testCase.Expected) {
			t.Errorf("Expected service labels: %v\ngot: %v", testCase.Expected, serviceReq.LocalService.ObjectMeta.Labels)
		}
	}
}
