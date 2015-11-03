// ThreatSpec package main
package main

import (
	"github.com/jawher/mow.cli"
	"github.com/pki-io/controllers/org"
)

func initCmd(cmd *cli.Cmd) {
	cmd.Spec = "ORG [OPTIONS]"

	params := org.NewParams()
	params.org = cmd.StringArg("ORG", "", "name of organization")
	params.admin = cmd.StringOpt("admin", "admin", "name of admin")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("initialising new org")

		cont, err := org.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Init(params); err != nil {
			app.Fatal(err)
		}
	}
}
