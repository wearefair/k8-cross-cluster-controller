package main

import (
	"testing"

	"k8s.io/client-go/rest"
)

func TestValidateK8Conf(t *testing.T) {
	testCases := []struct {
		local  *rest.Config
		remote *rest.Config
		err    error
	}{
		// If the hosts of the rest configs are the same, expect an error
		{
			local: &rest.Config{
				Host: "foo",
			},
			remote: &rest.Config{
				Host: "foo",
			},
			err: ErrLocalRemoteK8ConfMatch,
		},
		// If the hosts do not match, do not return an error
		{
			local: &rest.Config{
				Host: "bar",
			},
			remote: &rest.Config{
				Host: "baz",
			},
		},
	}

	for _, testCase := range testCases {
		err := validateK8Conf(testCase.local, testCase.remote)
		if err != testCase.err {
			t.Errorf("Expected error: %v\n, got %v", testCase.err, err)
		}
	}
}
