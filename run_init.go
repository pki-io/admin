// ThreatSpec package main
package main

import (
	"github.com/jawher/mow.cli"
	"github.com/pki-io/controller"
)

func initCmd(cmd *cli.Cmd) {
	cmd.Spec = "ORG [OPTIONS]"

	params := controller.NewOrgParams()
	params.Org = cmd.StringArg("ORG", "", "name of organization")
	params.Admin = cmd.StringOpt("admin", "admin", "name of admin")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("initialising new org")

		cont, err := controller.NewOrg(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Init(params); err != nil {
			app.Fatal(err)
		}
	}
}
