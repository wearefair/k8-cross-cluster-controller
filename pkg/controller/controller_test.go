package controller

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceWhitelist(t *testing.T) {
	originalSvc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
			Labels: map[string]string{
				"arf": "meow",
			},
			ClusterName:     "liz lemon",
			SelfLink:        "this better not make it in",
			ResourceVersion: "it's over 9000",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "so many ports",
				},
			},
			SessionAffinity: v1.ServiceAffinityNone,
			ClusterIP:       "127.0.0.1",
		},
	}
	testCases := []struct {
		NewService *v1.Service
		Expected   *v1.Service
	}{
		// Empty new service gets name, namespace, labels, clustername, ports, and session affinity copied over
		{
			NewService: &v1.Service{},
			Expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"arf": "meow",
					},
					ClusterName: "liz lemon",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Name: "so many ports",
						},
					},
					SessionAffinity: v1.ServiceAffinityNone,
				},
			},
		},
		// Service with values filled in gets name, namespace, labels, clustername, ports, and session affinity copied over
		// but retains original other values
		{
			NewService: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "blah",
					Namespace: "bah",
					Labels: map[string]string{
						"abc": "def",
					},
					ClusterName:     "not the same cluster",
					ResourceVersion: "this is totally a thing",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Name: "portugal the man",
						},
					},
					SessionAffinity: v1.ServiceAffinityNone,
					ExternalName:    "this will make it over",
				},
			},
			Expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"arf": "meow",
					},
					ClusterName:     "liz lemon",
					ResourceVersion: "this is totally a thing",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Name: "so many ports",
						},
					},
					SessionAffinity: v1.ServiceAffinityNone,
					ExternalName:    "this will make it over",
				},
			},
		},
	}

	for _, testCase := range testCases {
		out := serviceWhitelist(originalSvc, testCase.NewService)
		if !reflect.DeepEqual(out, testCase.Expected) {
			t.Errorf("Expected: %+v\ngot: %+v", testCase.Expected, out)
		}
	}
}

func TestEndpointsWhitelist(t *testing.T) {
	originalEndpoints := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
			Labels: map[string]string{
				"arf": "meow",
			},
			ClusterName:     "liz lemon",
			SelfLink:        "this better not make it in",
			ResourceVersion: "it's over 9000",
		},
		Subsets: []v1.EndpointSubset{
			v1.EndpointSubset{
				Addresses: []v1.EndpointAddress{
					v1.EndpointAddress{
						IP:       "totallylegitIPaddress",
						Hostname: "boopboop",
					},
				},
			},
		},
	}

	testCases := []struct {
		NewEndpoints *v1.Endpoints
		Expected     *v1.Endpoints
	}{
		// Empty endpoints gets name, namespace, labels, cluster name, and subsets copied over
		{
			NewEndpoints: &v1.Endpoints{},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"arf": "meow",
					},
					ClusterName: "liz lemon",
				},
				Subsets: []v1.EndpointSubset{
					v1.EndpointSubset{
						Addresses: []v1.EndpointAddress{
							v1.EndpointAddress{
								IP:       "totallylegitIPaddress",
								Hostname: "boopboop",
							},
						},
					},
				},
			},
		},
		// Endpoints will get name, namespace, labels, clustername, and subsets copied over, but retain other original values
		{
			NewEndpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "herp",
					Namespace: "derp",
					Labels: map[string]string{
						"fake": "label",
					},
					ClusterName:     "deepspacenine",
					SelfLink:        "blah",
					ResourceVersion: "sauce",
				},
				Subsets: []v1.EndpointSubset{
					v1.EndpointSubset{
						Addresses: []v1.EndpointAddress{
							v1.EndpointAddress{
								IP:       "anotherIPaddr",
								Hostname: "beepbeep",
							},
						},
					},
				},
			},
			Expected: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"arf": "meow",
					},
					ClusterName:     "liz lemon",
					SelfLink:        "blah",
					ResourceVersion: "sauce",
				},
				Subsets: []v1.EndpointSubset{
					v1.EndpointSubset{
						Addresses: []v1.EndpointAddress{
							v1.EndpointAddress{
								IP:       "totallylegitIPaddress",
								Hostname: "boopboop",
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		out := endpointsWhitelist(originalEndpoints, testCase.NewEndpoints)
		if !reflect.DeepEqual(out, testCase.Expected) {
			t.Errorf("Expected endpoints %+v\ngot: %+v", testCase.Expected, out)
		}
	}
}
