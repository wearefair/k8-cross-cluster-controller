package k8

import (
	"errors"
	"testing"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestResourceNotExist(t *testing.T) {
	fakeGroupResource := schema.GroupResource{
		Group:    "fake",
		Resource: "resource",
	}
	testCases := []struct {
		Err      error
		Expected bool
	}{
		// NotFound error returns true
		{
			Err:      kerrors.NewNotFound(fakeGroupResource, "foo"),
			Expected: true,
		},
		// IsGone error returns true
		{
			Err:      kerrors.NewGone("foo"),
			Expected: true,
		},
		// K8 error that isn't a IsGone or NotFound error returns false
		{
			Err: kerrors.NewUnauthorized("oh no"),
		},
		// Non-K8 error returns false
		{
			Err: errors.New("oh the errors"),
		},
		// Nil returns false
		{},
	}

	for _, testCase := range testCases {
		res := ResourceNotExist(testCase.Err)
		if res != testCase.Expected {
			t.Errorf("Expected %t, but got %t", testCase.Expected, res)
		}
	}
}
