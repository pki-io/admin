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

	exists, err := app.AdminConfigExists()
	checkAppFatal("Could not check admin config existence: %s", err)

	if exists {
		logger.Info("Existing admin config found")
		app.LoadAdminConfig()
	} else {
		app.CreateAdminConfig()
	}

	if app.entities.admin == nil {
		checkAppFatal("admin entity cannot be nil")
	}

	err = app.config.admin.AddOrg(app.config.org.Data.Name, app.config.org.Data.Id, app.entities.admin.Data.Body.Id)
	checkUserFatal("Cannot add org to admin config: %s", err)

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
