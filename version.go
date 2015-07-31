package main

import (
	"fmt"
)

const (
	name    string = "pki.io"
	version string = "0.2.2"
	release string = "development"
)

func Version() string {
	return fmt.Sprintf("%s %s-%s", name, version, release)
}
