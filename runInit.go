// ThreatSpec package main
package main

import (
	"github.com/jawher/mow.cli"
)

func initCmd(cmd *cli.Cmd) {
	cmd.Spec = "ORG [OPTIONS]"

	params := NewOrgParams()
	params.org = cmd.StringArg("ORG", "", "name of organization")
	params.admin = cmd.StringOpt("admin", "admin", "name of admin")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("initialising new org")

		cont, err := NewOrgController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Init(params); err != nil {
			app.Fatal(err)
		}
	}
}
