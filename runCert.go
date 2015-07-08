package main

import (
	"crypto/x509/pkix"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/x509"
)

func certNew(argv map[string]interface{}) (err error) {
	// TODO - this whole function needs to be refactored
	name := ArgString(argv["<name>"], nil)
	exportFile := ArgString(argv["--export"], nil)
	expiry := ArgInt(argv["--expiry"], 365)

	caName := ArgString(argv["--ca"], "")

	dnLocality := ArgString(argv["--dn-l"], "")
	dnState := ArgString(argv["--dn-st"], "")
	dnOrg := ArgString(argv["--dn-o"], "")
	dnOrgUnit := ArgString(argv["--dn-ou"], "")
	dnCountry := ArgString(argv["--dn-c"], "")
	dnStreet := ArgString(argv["--dn-street"], "")
	dnPostal := ArgString(argv["--dn-postal"], "")

	// TODO - This should really be in a certificate function
	subject := pkix.Name{CommonName: name}

	if dnLocality != "" {
		subject.Locality = []string{dnLocality}
	}
	if dnState != "" {
		subject.Province = []string{dnState}
	}
	if dnOrg != "" {
		subject.Organization = []string{dnOrg}
	}
	if dnOrgUnit != "" {
		subject.OrganizationalUnit = []string{dnOrgUnit}
	}
	if dnCountry != "" {
		subject.Country = []string{dnCountry}
	}
	if dnStreet != "" {
		subject.StreetAddress = []string{dnStreet}
	}
	if dnPostal != "" {
		subject.PostalCode = []string{dnPostal}
	}

	cert, err := x509.NewCertificate(nil)
	checkAppFatal("Couldn't create new certificate: %s", err)

	cert.Data.Body.Name = name
	cert.Data.Body.Expiry = expiry

	var files []ExportFile
	certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
	keyFile := fmt.Sprintf("%s-key.pem", cert.Data.Body.Name)

	caFile := fmt.Sprintf("%s-cacert.pem", cert.Data.Body.Name)
	if caName == "" {
		// Self-signed
		err := cert.Generate(nil, &subject)
		checkAppFatal("Couldn't generate certificate: %s", err)

		files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
	} else {
		app := NewAdminApp()
		app.Load()
		app.LoadOrgIndex()

		caId, err := app.index.org.GetCA(caName)
		checkUserFatal("Couldn't find CA '%s'", caName)

		caContainerJson, err := app.fs.api.GetPrivate(app.entities.org.Data.Body.Id, caId)
		caContainer, err := document.NewContainer(caContainerJson)
		checkAppFatal("Couldn't create container from json: %s", err)

		caJson, err := app.entities.org.VerifyThenDecrypt(caContainer)
		checkAppFatal("Couldn't verify and decrypt ca container: %s", err)

		ca, err := x509.NewCA(caJson)
		checkAppFatal("Couldn't create ca: %s", err)

		err = cert.Generate(ca, &subject)
		checkAppFatal("Couldn't generate certificate: %s", err)
		files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(ca.Data.Body.Certificate)})
	}

	files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
	files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})

	if caName == "" {
	} else {
	}

	logger.Infof("Export to '%s'", exportFile)
	Export(files, exportFile)

	return nil
}

func runCert(args []string) (err error) {
	usage := `
Manages standalone certificates

Usage: 
    pki.io cert [--help]
    pki.io cert new <name> --export <file> [--expiry <days>] [--ca <ca>] [--dn-l <locality>] [--dn-st <state>] [--dn-o <org>] [--dn-ou <orgUnit>] [--dn-c <country>] [--dn-street <street>] [--dn-postal <postalCode>]

Options:
    --export <file>  Export certificate to <file>
    --expiry <days>  Expiry period in days [default: 365]
    --ca <ca>        Name of CA
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
		certNew(argv)
	}
	return nil
}
