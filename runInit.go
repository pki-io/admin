package main

import (
	"github.com/docopt/docopt-go"
)

func newInit(argv map[string]interface{}) {
	var adminName string
	orgName := argv["<org>"].(string)
	if argv["--admin"] == nil {
		adminName = "admin"
	} else {
		adminName = argv["--admin"].(string)
	}

	// TODO - Use the AdminApp

	app := NewAdminApp()

	app.InitLocalFs()
	app.CreateOrgDirectory(orgName)

	app.InitApiFs()
	app.InitHomeFs()

	app.CreateAdminEntity(adminName)
	app.CreateOrgEntity(orgName)

	app.SaveAdminEntity()
	app.CreateAdminConfig()

	app.SaveOrgEntity()
	app.CreateOrgConfig()

	app.CreateOrgIndex()
	app.SaveOrgIndex()

	app.config.org.Data.Index = app.index.org.Data.Body.Id
	app.SaveAdminConfig()
	app.SaveOrgConfig()
}

func runInit(args []string) (err error) {

	usage := `
Initialises a new Organisation.

Usage: pki.io init <org> [--admin=<admin>]

Options
    --admin   Admin name. Defaults to 'admin'
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	newInit(argv)
	return nil
}
