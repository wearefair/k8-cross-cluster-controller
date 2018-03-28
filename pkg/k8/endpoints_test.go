package k8

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEndpointsWriterUpdate(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	testCases := []struct {
		EndpointToCreate *v1.Endpoints
		UpdateEndpoint   *v1.Endpoints
		ExpectedEndpoint *v1.Endpoints
	}{
		// Endpoints object already exists, update
		{
			EndpointToCreate: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"wow": "solabel",
					},
				},
			},
			UpdateEndpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
			ExpectedEndpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
		},
		// Endpoints object doesn't exist, attempt a create if update fails
		{
			UpdateEndpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
			ExpectedEndpoint: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		writer := &EndpointsWriter{
			Events: make(chan *EndpointsRequest, 4),
			Client: fakeClientSet,
		}
		if testCase.EndpointToCreate != nil {
			_, err := fakeClientSet.CoreV1().Endpoints(testCase.EndpointToCreate.ObjectMeta.Namespace).Create(testCase.EndpointToCreate)
			if err != nil {
				t.Fatalf("Something went wrong creating fake endpoint against fake clientset %v", err)
			}
		}
		writer.update(testCase.UpdateEndpoint)
		endpoints, err := fakeClientSet.CoreV1().
			Endpoints(testCase.ExpectedEndpoint.ObjectMeta.Namespace).
			Get(testCase.ExpectedEndpoint.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			t.Errorf("Could not get endpoint %v", err)
		}
		if !reflect.DeepEqual(testCase.ExpectedEndpoint, endpoints) {
			t.Errorf("Expected endpoint: %+v\ngot: %+v", testCase.ExpectedEndpoint, endpoints)
		}
	}
}
