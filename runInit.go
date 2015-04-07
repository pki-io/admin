package main

import (
	"github.com/docopt/docopt-go"
)

func newInit(argv map[string]interface{}) {
	orgName := ArgString(argv["<org>"], nil)
	adminName := ArgString(argv["--admin"], "admin")

	app := NewAdminApp()

	app.InitLocalFs()
	app.CreateOrgDirectory(orgName)

	app.InitApiFs()
	app.InitHomeFs()

	app.CreateAdminEntity(adminName)
	app.CreateOrgEntity(orgName)
	app.CreateOrgConfig()

	app.SaveAdminEntity()

	app.CreateOrgIndex()
	err := app.index.org.AddAdmin(app.entities.admin.Data.Body.Name, app.entities.admin.Data.Body.Id)
	checkAppFatal("Couldn't add admin to index: %s", err)

	app.SaveOrgIndex()

	app.SaveOrgEntityPublic()
	app.CreateAdminConfig()

	app.SendOrgEntity()

	app.config.org.Data.Index = app.index.org.Data.Body.Id
	app.SaveAdminConfig()
	app.SaveOrgConfig()
}

func runInit(args []string) (err error) {

	usage := `
Initialises a new Organisation.

Usage: pki.io init <org> [--admin <admin>]

Options
    --admin <admin>  Admin name. Defaults to 'admin'
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	newInit(argv)
	return nil
}
