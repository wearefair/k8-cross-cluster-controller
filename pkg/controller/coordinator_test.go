package controller

import (
	"reflect"
	"testing"

	"github.com/wearefair/k8-cross-cluster-controller/pkg/k8"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplyCrossClusterLabelToEndpoints(t *testing.T) {
	testCases := []struct {
		Endpoint *v1.Endpoints
		Expected *v1.Endpoints
	}{
		// Endpoint with labels, but not the cross cluster label gets it added
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
						"baz": "yo",
					},
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
						"baz": "yo",
						k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLabelValue,
					},
				},
			},
		},
		// Endpoint with cross cluster label already doesn't get affected
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLabelValue,
					},
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLabelValue,
					},
				},
			},
		},
		// Endpoint with cross cluster label and other labels doesn't get affected
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
						k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLabelValue,
					},
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
						k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLabelValue,
					},
				},
			},
		},
		// Endpoint with no labels gets the cross cluster label added
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLabelValue,
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		applyCrossClusterLabelToEndpoints(testCase.Endpoint)
		if !reflect.DeepEqual(testCase.Endpoint, testCase.Expected) {
			t.Errorf("Expected: %+v\ngot: %+v", testCase.Expected, testCase.Endpoint)
		}
	}
}
