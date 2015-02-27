package main

import (
	"github.com/docopt/docopt-go"
)

func orgShow(argv map[string]interface{}) (err error) {
	app := NewAdminApp()
	app.Load()

	app.LoadOrgIndex()
	logger.Info(app.index.org)
	logger.Info(app.entities.org)
	return nil
}

func orgRun(argv map[string]interface{}) (err error) {
	app := NewAdminApp()
	app.Load()

	app.LoadOrgIndex()
	app.RegisterNodes()
	app.SaveOrgIndex()
	return nil
}

func runOrg(args []string) (err error) {
	usage := `
Manages the Organisation.

Usage:
    pki.io org [--help]
    pki.io org run
    pki.io org show
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["show"].(bool) {
		orgShow(argv)
	} else if argv["run"].(bool) {
		orgRun(argv)
	}
	return nil
}
