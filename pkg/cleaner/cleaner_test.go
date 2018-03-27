package cleaner

import (
	"errors"
	"testing"

	"k8s.io/api/apps/v1beta1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8ResourceDoesNotExist(t *testing.T) {
	fakeGroup := schema.GroupResource{
		Group:    v1beta1.SchemeGroupVersion.Group,
		Resource: "blah",
	}
	testCases := []struct {
		Err      error
		Expected bool
	}{
		// K8 error that isn't a Gone/NotFound type returns false
		{
			Err: k8errors.NewBadRequest("foo"),
		},
		// K8 error that is a Gone type returns true
		{
			Err:      k8errors.NewGone("baby gone"),
			Expected: true,
		},
		// K8 error that is a NotFound type returns true
		{
			Err:      k8errors.NewNotFound(fakeGroup, "404"),
			Expected: true,
		},
		// Non-K8 error type returns false
		{
			Err: errors.New("oh no"),
		},
		// Nil returns false
		{},
	}

	for _, testCase := range testCases {
		res := k8ResourceDoesNotExist(testCase.Err)
		if testCase.Expected != res {
			t.Errorf("Expected %t, got %t", testCase.Expected, res)
		}
	}
}
