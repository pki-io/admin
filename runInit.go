// ThreatSpec package main
package main

import (
	"github.com/docopt/docopt-go"
)

// ThreatSpec TMv0.1 for newInit
// Does new org initialisation for App:Org

func newInit(argv map[string]interface{}) {
	orgName := ArgString(argv["<org>"], nil)
	adminName := ArgString(argv["--admin"], "admin")

	app := NewAdminApp()

	app.InitLocalFs()
	app.InitHomeFs()

	// Do some checks now to prevent changes before obvious errors
	app.ErrorIfOrgDirectoryExists(orgName)
	// TODO - check admin config for org and error if it exists

	exists, err := app.AdminConfigExists()
	checkAppFatal("Could not check admin config existence: %s", err)

	if exists {
		logger.Info("Existing admin config found")
		app.LoadAdminConfig()
	} else {
		app.CreateAdminConfig()
	}

	if app.config.admin.OrgExists(orgName) {
		checkUserFatal("Org '%s' already exists in admin config.", orgName)
	}

	app.CreateOrgDirectory(orgName)

	// Initialise the API fs asap but it depends on an org config existing
	app.InitApiFs()

	app.CreateAdminEntity(adminName)
	app.CreateOrgEntity(orgName)
	app.CreateOrgConfig()

	app.SaveAdminEntity()

	app.CreateOrgIndex()
	err = app.index.org.AddAdmin(app.entities.admin.Data.Body.Name, app.entities.admin.Data.Body.Id)
	checkAppFatal("Couldn't add admin to index: %s", err)

	app.SaveOrgIndex()

	app.SaveOrgEntityPublic()

	if app.entities.admin == nil {
		checkAppFatal("admin entity cannot be nil")
	}

	err = app.config.admin.AddOrg(app.config.org.Data.Name, app.config.org.Data.Id, app.entities.admin.Data.Body.Id)
	checkUserFatal("Could not add org to admin config: %s", err)

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
