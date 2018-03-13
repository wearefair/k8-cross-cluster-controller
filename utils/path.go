package utils

import (
	"context"
	"path/filepath"

	"github.com/wearefair/service-kit-go/errors"
)

// PathHelper takes a path and will attempt to get the absolute path
func PathHelper(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	finalPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", errors.Error(context.Background(), err)
	}
	return finalPath, nil
}
