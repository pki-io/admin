package main

import (
	"github.com/docopt/docopt-go"
	"github.com/pki-io/core/x509"
)

func caNew(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	inTags := ArgString(argv["--tags"], nil)

	caExpiry := ArgInt(argv["--ca-expiry"], 365)
	certExpiry := ArgInt(argv["--cert-expiry"], 90)

	dnLocality := ArgString(argv["--dn-l"], "")
	dnState := ArgString(argv["--dn-st"], "")
	dnOrg := ArgString(argv["--dn-o"], "")
	dnOrgUnit := ArgString(argv["--dn-ou"], "")
	dnCountry := ArgString(argv["--dn-c"], "")
	dnStreet := ArgString(argv["--dn-street"], "")
	dnPostal := ArgString(argv["--dn-postal"], "")

	app := NewAdminApp()
	app.Load()

	ca, _ := x509.NewCA(nil)
	ca.Data.Body.Name = name
	ca.Data.Body.CAExpiry = caExpiry
	ca.Data.Body.CertExpiry = certExpiry

	if dnLocality != "" {
		ca.Data.Body.DNScope.Locality = dnLocality
	}
	if dnState != "" {
		ca.Data.Body.DNScope.Province = dnState
	}
	if dnOrg != "" {
		ca.Data.Body.DNScope.Organization = dnOrg
	}
	if dnOrgUnit != "" {
		ca.Data.Body.DNScope.OrganizationalUnit = dnOrgUnit
	}
	if dnCountry != "" {
		ca.Data.Body.DNScope.Country = dnCountry
	}
	if dnStreet != "" {
		ca.Data.Body.DNScope.StreetAddress = dnStreet
	}
	if dnPostal != "" {
		ca.Data.Body.DNScope.PostalCode = dnPostal
	}

	ca.GenerateRoot()

	logger.Info("Saving CA")
	caContainer, err := app.entities.org.EncryptThenSignString(ca.Dump(), nil)
	checkAppFatal("Could not encrypt CA: %s", err)

	err = app.fs.api.SendPrivate(app.entities.org.Data.Body.Id, ca.Data.Body.Id, caContainer.Dump())
	checkAppFatal("Could not save CA: %s", err)

	logger.Info("Updating index")
	app.LoadOrgIndex()
	app.index.org.AddCA(ca.Data.Body.Name, ca.Data.Body.Id)
	app.index.org.AddCATags(ca.Data.Body.Id, ParseTags(inTags))
	app.SaveOrgIndex()

	return nil
}

func runCA(args []string) (err error) {
	usage := `
Manages Certificate Authorities

Usage: 
    pki.io ca [--help]
    pki.io ca new <name> --tags <tags> [--ca-expiry <days>] [--cert-expiry <days>] [--dn-l <locality>] [--dn-st <state>] [--dn-o <org>] [--dn-ou <orgUnit>] [--dn-c <country>] [--dn-street <street>] [--dn-postal <postalCode>]

Options:
    --tags <tags>             List of comma-separated tags
    --ca-expiry <days>        Expiry period for CA in days [default: 365]
    --cert-expiry <days>      Expiry period for certs in day [default: 90]
    --dn-l <locality>         Locality for DN scope
    --dn-st <state>           State/province for DN scope
    --dn-o <org>              Organization for DN scope
    --dn-ou <orgUnit>         Organizational unit for DN scope
    --dn-c <country>          Country for DN scope
    --dn-street <street>      Street for DN scope
    --dn-postal <postalCode>  Postal code for DN scope
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["new"].(bool) {
		caNew(argv)
	}
	return nil
}
