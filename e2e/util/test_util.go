package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectDir returns the root location of the project based on the GOPATH env variable.
func ProjectDir() string {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/sky-uk/fluentd-docker")
}

func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func Resource(name string) string {
	return fmt.Sprintf("%s/%s", ProjectDir(), name)
}


