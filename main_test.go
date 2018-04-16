package main

import (
	"testing"

	"k8s.io/client-go/rest"
)

func TestValidateK8Conf(t *testing.T) {
	testCases := []struct {
		local   *rest.Config
		remote  *rest.Config
		devMode bool
		err     error
	}{
		// If dev mode is disabled and the hosts of the rest configs are
		// the same, expect an error returned
		{
			local: &rest.Config{
				Host: "foo",
			},
			remote: &rest.Config{
				Host: "foo",
			},
			err: ErrLocalRemoteK8ConfMatch,
		},
		// If dev mode is enabled and the hosts of the rest configs are the same
		// do not return an error
		{
			local: &rest.Config{
				Host: "foo",
			},
			remote: &rest.Config{
				Host: "foo",
			},
			devMode: true,
		},
		// If dev mode is not enabled and the hosts do not match, do not return an error
		{
			local: &rest.Config{
				Host: "bar",
			},
			remote: &rest.Config{
				Host: "baz",
			},
		},
		// If dev mode is enabled and the hosts do not match, do not return an error
		{
			local: &rest.Config{
				Host: "bar",
			},
			remote: &rest.Config{
				Host: "baz",
			},
			devMode: true,
		},
	}

	for _, testCase := range testCases {
		if testCase.devMode {
			devMode = "true"
			defer func() {
				devMode = "false"
			}()
		}
		err := validateK8Conf(testCase.local, testCase.remote)
		if err != testCase.err {
			t.Errorf("Expected error: %v\n, got %v", testCase.err, err)
		}
	}
}
