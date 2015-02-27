package main

import (
	"github.com/docopt/docopt-go"
	"github.com/pki-io/pki.io/x509"
	"time"
)

func caNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)
	inTags := argv["--tags"].(string)

	app := NewAdminApp()
	app.Load()

	ca, _ := x509.NewCA(nil)
	ca.Data.Body.Name = name
	ca.GenerateRoot(time.Now(), time.Now().AddDate(5, 5, 5))

	logger.Info("Saving CA")
	caContainer, err := app.entities.org.EncryptThenSignString(ca.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt CA: %s", err))
	}

	if err := app.fs.api.SendPrivate(app.entities.org.Data.Body.Id, ca.Data.Body.Id, caContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save CA: %s", err))
	}

	logger.Info("Updating index")
	app.LoadOrgIndex()
	app.index.org.AddCATags(ca.Data.Body.Id, ParseTags(inTags))
	app.SaveOrgIndex()

	return nil
}

func runCA(args []string) (err error) {
	usage := `
Manages Certificate Authorities

Usage:
    pki.io ca [--help]
    pki.io ca new <name> --tags=<tags>

Options:
    --tags=<tags>   List of comma-separated tags
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["new"].(bool) {
		caNew(argv)
	}
	return nil
}
