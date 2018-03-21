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

func TestSanitizeEndpointsResourceVersion(t *testing.T) {
	testCases := []struct {
		Endpoint *v1.Endpoints
		Expected *v1.Endpoints
	}{
		// Endpoint without resource version has it set to empty string
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "",
				},
			},
		},
		// Endpoint with resource version has it set to empty string
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "foo",
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "",
				},
			},
		},
		// Endpoint with empty resource version has it set to empty string
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "",
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "",
				},
			},
		},
	}

	for _, testCase := range testCases {
		sanitizeEndpointsResourceVersion(testCase.Endpoint)
		if !reflect.DeepEqual(testCase.Endpoint, testCase.Expected) {
			t.Errorf("Expected: %+v\ngot: %+v", testCase.Expected, testCase.Endpoint)
		}
	}
}

func TestSanitizeEndpointsUID(t *testing.T) {
	testCases := []struct {
		Endpoint *v1.Endpoints
		Expected *v1.Endpoints
	}{
		// Endpoint without UID has it set to empty string
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					UID: "",
				},
			},
		},
		// Endpoint with UID has it set to empty string
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					UID: "foo",
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					UID: "",
				},
			},
		},
		// Endpoint with empty UID has it set to empty string
		{
			Endpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					UID: "",
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					UID: "",
				},
			},
		},
	}

	for _, testCase := range testCases {
		sanitizeEndpointsUID(testCase.Endpoint)
		if !reflect.DeepEqual(testCase.Endpoint, testCase.Expected) {
			t.Errorf("Expected: %+v\ngot: %+v", testCase.Expected, testCase.Endpoint)
		}
	}
}

func TestApplyEndpointsTransformations(t *testing.T) {
	input := &v1.Endpoints{}
	expected := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "",
			Labels: map[string]string{
				k8.CrossClusterServiceLabelKey: k8.CrossClusterServiceLabelValue,
			},
		},
	}

	// Tests to ensure that endpoint transformers chain appropriately
	applyEndpointsTransformations(input, sanitizeEndpointsResourceVersion, applyCrossClusterLabelToEndpoints)

	if !reflect.DeepEqual(input, expected) {
		t.Errorf("Expected: %+v\ngot: %+v", expected, input)
	}
}

func TestSanitizeServiceResourceVersion(t *testing.T) {
	testCases := []struct {
		Service  *v1.Service
		Expected *v1.Service
	}{
		// Service without resource version has it set to empty string
		{
			Service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{},
			},
			Expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "",
				},
			},
		},
		// Service with resource version has it set to empty string
		{
			Service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "It's after six. What am I, a farmer?",
				},
			},
			Expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "",
				},
			},
		},
		// Service with empty resource version has it set to empty string
		{
			Service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "",
				},
			},
			Expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "",
				},
			},
		},
	}

	for _, testCase := range testCases {
		sanitizeServiceResourceVersion(testCase.Service)
		if !reflect.DeepEqual(testCase.Service, testCase.Expected) {
			t.Errorf("Expected: %+v\ngot: %+v", testCase.Expected, testCase.Service)
		}
	}
}

func TestSanitizeServiceClusterIP(t *testing.T) {
	testCases := []struct {
		Service  *v1.Service
		Expected *v1.Service
	}{
		// Service without cluster IP has it set to empty string
		{
			Service: &v1.Service{
				Spec: v1.ServiceSpec{
					ClusterIP: "",
				},
			},
			Expected: &v1.Service{
				Spec: v1.ServiceSpec{
					ClusterIP: "",
				},
			},
		},
		// Service with clusterIP has it set to empty string
		{
			Service: &v1.Service{
				Spec: v1.ServiceSpec{
					ClusterIP: "Wow this is totally a legit IP",
				},
			},
			Expected: &v1.Service{
				Spec: v1.ServiceSpec{
					ClusterIP: "",
				},
			},
		},
		// Service with empty cluster IP has it set to empty string
		{
			Service: &v1.Service{
				Spec: v1.ServiceSpec{
					ClusterIP: "",
				},
			},
			Expected: &v1.Service{
				Spec: v1.ServiceSpec{
					ClusterIP: "",
				},
			},
		},
	}

	for _, testCase := range testCases {
		sanitizeServiceClusterIP(testCase.Service)
		if !reflect.DeepEqual(testCase.Service, testCase.Expected) {
			t.Errorf("Expected: %+v\ngot: %+v", testCase.Expected, testCase.Service)
		}
	}
}

func TestApplyServiceTransformations(t *testing.T) {
	input := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "meow meow",
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "ipv6 lolz",
		},
	}
	expected := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "",
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "",
		},
	}

	// Tests to ensure that service transformers chain appropriately
	applyServiceTransformations(input, sanitizeServiceClusterIP, sanitizeServiceResourceVersion)

	if !reflect.DeepEqual(input, expected) {
		t.Errorf("Expected: %+v\ngot: %+v", expected, input)
	}
}
