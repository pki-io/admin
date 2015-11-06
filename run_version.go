// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"os"
)

const (
	Version string = "0.4.0"
	Release string = "development"
)

func versionCmd(cmd *cli.Cmd) {
	cmd.Action = func() {
		fmt.Printf("%s-%s\n", Version, Release)
		os.Exit(0)
	}
}
