package main

import (
	"fmt"
)

const (
	name    string = "pki.io"
	version string = "0.2.0"
	release string = "release1"
)

func Version() string {
	return fmt.Sprintf("%s %s-%s", name, version, release)
}
