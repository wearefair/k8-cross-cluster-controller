package cleaner

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResourceExists(t *testing.T) {
	testCases := []struct {
		LocalMeta  metav1.ObjectMeta
		RemoteMeta metav1.ObjectMeta
		Expected   bool
	}{
		// If local meta and remote meta are empty, return true
		{
			LocalMeta:  metav1.ObjectMeta{},
			RemoteMeta: metav1.ObjectMeta{},
			Expected:   true,
		},
		// If local meta and remote meta have the same name, but different namespaces, return false
		{
			LocalMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
			RemoteMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "baz",
			},
		},
		// If local meta and remote meta have the same namespace, but different names, return false
		{
			LocalMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "foo",
			},
			RemoteMeta: metav1.ObjectMeta{
				Name:      "baz",
				Namespace: "foo",
			},
		},
		// If local meta and remote meta have the same name and same namespace, return true
		{
			LocalMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "foo",
			},
			RemoteMeta: metav1.ObjectMeta{
				Name:      "bar",
				Namespace: "foo",
			},
			Expected: true,
		},
	}
	for _, testCase := range testCases {
		res := resourceExists(testCase.LocalMeta, testCase.RemoteMeta)
		if testCase.Expected != res {
			t.Errorf("Expected: %t, got %t", testCase.Expected, res)
		}
	}
}
