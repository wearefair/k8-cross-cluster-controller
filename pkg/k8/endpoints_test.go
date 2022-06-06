package k8

import (
	"context"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEndpointsWriterUpdate(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	testCases := []struct {
		EndpointsToCreate *v1.Endpoints
		UpdateEndpoints   *v1.Endpoints
		ExpectedEndpoints *v1.Endpoints
	}{
		// Endpoints object already exists, update
		{
			EndpointsToCreate: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"wow": "solabel",
					},
				},
			},
			UpdateEndpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
			ExpectedEndpoints: &v1.Endpoints{
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
			UpdateEndpoints: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
			ExpectedEndpoints: &v1.Endpoints{
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
		ctx := context.TODO()
		writer := &EndpointsWriter{
			Events: make(chan *EndpointsRequest, 4),
			Client: fakeClientSet,
		}
		if testCase.EndpointsToCreate != nil {
			_, err := fakeClientSet.CoreV1().Endpoints(testCase.EndpointsToCreate.ObjectMeta.Namespace).Create(ctx, testCase.EndpointsToCreate, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("Something went wrong creating fake endpoint against fake clientset %v", err)
			}
		}
		writer.update(ctx, testCase.UpdateEndpoints)
		endpoints, err := fakeClientSet.CoreV1().
			Endpoints(testCase.ExpectedEndpoints.ObjectMeta.Namespace).
			Get(ctx, testCase.ExpectedEndpoints.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			t.Errorf("Could not get endpoint %v", err)
		}
		if !reflect.DeepEqual(testCase.ExpectedEndpoints, endpoints) {
			t.Errorf("Expected endpoint: %+v\ngot: %+v", testCase.ExpectedEndpoints, endpoints)
		}
	}
}
