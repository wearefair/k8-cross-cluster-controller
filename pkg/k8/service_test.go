package k8

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServicesWriterUpdate(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset()
	testCases := []struct {
		ServiceToCreate *v1.Service
		UpdateService   *v1.Service
		ExpectedService *v1.Service
	}{
		// Service object already exists, update
		{
			ServiceToCreate: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"wow": "solabel",
					},
				},
			},
			UpdateService: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
			ExpectedService: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
		},
		// Service object doesn't exist, attempt a create if update fails
		{
			UpdateService: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "dead space",
					Labels: map[string]string{
						"not": "solabel",
					},
				},
			},
			ExpectedService: &v1.Service{
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
		writer := &ServiceWriter{
			Events: make(chan *ServiceRequest, 4),
			Client: fakeClientSet,
		}
		if testCase.ServiceToCreate != nil {
			_, err := fakeClientSet.CoreV1().Services(testCase.ServiceToCreate.ObjectMeta.Namespace).Create(testCase.ServiceToCreate)
			if err != nil {
				t.Fatalf("Something went wrong creating fake service against fake clientset %v", err)
			}
		}
		writer.update(testCase.UpdateService)
		service, err := fakeClientSet.CoreV1().
			Services(testCase.ExpectedService.ObjectMeta.Namespace).
			Get(testCase.ExpectedService.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			t.Errorf("Could not get service %v", err)
		}
		if !reflect.DeepEqual(testCase.ExpectedService, service) {
			t.Errorf("Expected service: %+v\ngot: %+v", testCase.ExpectedService, service)
		}
	}
}
