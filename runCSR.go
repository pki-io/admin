package main

import (
	"crypto/x509/pkix"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/fs"
	"github.com/pki-io/core/x509"
)

func csrNew(argv map[string]interface{}) (err error) {

	name := ArgString(argv["<name>"], nil)
	tags := ArgString(argv["--tags"], nil)

	standaloneFile := ArgString(argv["--standalone"], "")

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

	app := NewAdminApp()
	app.Load()

	logger.Info("Creating new CSR")
	csr, err := x509.NewCSR(nil)
	checkAppFatal("Could not generate CSR: %s", err)

	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = name
	csr.Generate(&subject)

	if standaloneFile == "" {
		logger.Info("Saving CSR")
		csrContainer, err := app.entities.org.EncryptThenSignString(csr.Dump(), nil)
		checkAppFatal("Could not encrypt CSR: %s", err)

		err = app.fs.api.Authenticate(app.entities.org.Data.Body.Id, "")
		checkAppFatal("Could not authenticate to API as Org: %s", err)

		err = app.fs.api.StorePrivate(csr.Data.Body.Id, csrContainer.Dump())
		checkAppFatal("Could not save cert: %s", err)

		logger.Info("Updating index")
		app.LoadOrgIndex()
		app.index.org.AddCSR(csr.Data.Body.Name, csr.Data.Body.Id)
		app.index.org.AddCSRTags(csr.Data.Body.Id, ParseTags(tags))
		app.SaveOrgIndex()
	} else {
		var files []ExportFile
		csrFile := fmt.Sprintf("%s-csr.pem", csr.Data.Body.Name)
		keyFile := fmt.Sprintf("%s-key.pem", csr.Data.Body.Name)
		files = append(files, ExportFile{Name: csrFile, Mode: 0644, Content: []byte(csr.Data.Body.CSR)})
		files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(csr.Data.Body.PrivateKey)})
		logger.Infof("Export to '%s'", standaloneFile)
		Export(files, standaloneFile)
	}

	return nil
}

func csrList(argv map[string]interface{}) (err error) {
	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()

	logger.Info("CSRs:")
	logger.Flush()
	for name, id := range app.index.org.GetCSRs() {
		fmt.Printf("* %s %s\n", name, id)
	}
	return nil
}

func csrShow(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	exportFile := ArgString(argv["--export"], "")
	private := ArgBool(argv["--private"], false)

	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()

	csrSerial, err := app.index.org.GetCSR(name)
	checkUserFatal("Could not find csr: %s%.0s\n", name, err)

	csr := app.GetCSR(csrSerial)

	if exportFile == "" {
		// TODO - show CA name
		fmt.Printf("Name: %s\n", csr.Data.Body.Name)
		fmt.Printf("ID: %s\n", csr.Data.Body.Id)
		fmt.Printf("Key type: %s\n", csr.Data.Body.KeyType)
		fmt.Printf("Certficate:\n%s\n", csr.Data.Body.CSR)

		if private {
			fmt.Printf("Private key:\n%s\n", csr.Data.Body.PrivateKey)
		}
	} else {
		var files []ExportFile
		csrFile := fmt.Sprintf("%s-csr.pem", csr.Data.Body.Name)
		keyFile := fmt.Sprintf("%s-key.pem", csr.Data.Body.Name)

		files = append(files, ExportFile{Name: csrFile, Mode: 0644, Content: []byte(csr.Data.Body.CSR)})

		if private {
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(csr.Data.Body.PrivateKey)})
		}

		logger.Infof("Export to '%s'", exportFile)
		Export(files, exportFile)
	}

	return nil
}

func csrDelete(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	reason := ArgString(argv["--confirm-delete"], nil)

	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()
	logger.Infof("Deleting csr %s with reason: %s", name, reason)

	csrId, err := app.index.org.GetCSR(name)
	checkUserFatal("csr %s does not exist%.0s", name, err)

	err = app.fs.api.Authenticate(app.entities.org.Data.Body.Id, "")
	checkAppFatal("Could not authenticate to API as Org: %s", err)

	err = app.fs.api.DeletePrivate(csrId)
	checkAppFatal("Could not delete CSR: %s", err)

	err = app.index.org.RemoveCSR(name)
	checkAppFatal("Could not remove CSR: %s", err)
	app.SaveOrgIndex()
	return nil
}

func csrSign(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	csrFile := ArgString(argv["<csrFile>"], nil)
	caName := ArgString(argv["--ca"], nil)
	tags := ArgString(argv["--tags"], nil)
	standaloneFile := ArgString(argv["--standalone"], "")

	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()

	ok, err := fs.Exists(csrFile)
	checkAppFatal("Could not check file existence for %s: %s", csrFile, err)
	if !ok {
		checkUserFatal("File does not exist: %s", csrFile)
	}
	csrPem, err := fs.ReadFile(csrFile)

	caId, err := app.index.org.GetCA(caName)
	checkUserFatal("Couldn't find CA '%s'%.0s", caName, err)

	// TODO - further validation on CSR
	_, err = x509.PemDecodeX509CSR([]byte(csrPem))
	checkUserFatal("Not a valid certificate request PEM for %s: %s", csrFile, err)

	csr, err := x509.NewCSR(nil)
	checkAppFatal("Couldn't create CSR: %s", err)

	// Use a random ID as we don't control the serial
	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = name
	csr.Data.Body.CSR = csrPem

	caContainerJson, err := app.fs.api.GetPrivate(app.entities.org.Data.Body.Id, caId)
	caContainer, err := document.NewContainer(caContainerJson)
	checkAppFatal("Couldn't create container from json: %s", err)

	caJson, err := app.entities.org.VerifyThenDecrypt(caContainer)
	checkAppFatal("Couldn't verify and decrypt ca container: %s", err)

	ca, err := x509.NewCA(caJson)
	checkAppFatal("Couldn't create ca: %s", err)

	logger.Info("Creating certificate")
	cert, err := ca.Sign(csr)
	checkAppFatal("Couldn't sign csr: %s", err)

	if standaloneFile == "" {
		logger.Info("Saving cert")
		certContainer, err := app.entities.org.EncryptThenSignString(cert.Dump(), nil)
		checkAppFatal("Could not encrypt CA: %s", err)

		err = app.fs.api.Authenticate(app.entities.org.Data.Body.Id, "")
		checkAppFatal("Could not authenticate to API as Org: %s", err)

		err = app.fs.api.StorePrivate(cert.Data.Body.Id, certContainer.Dump())
		checkAppFatal("Could not save cert: %s", err)

		logger.Info("Updating index")
		app.LoadOrgIndex()
		app.index.org.AddCert(cert.Data.Body.Name, cert.Data.Body.Id)
		app.index.org.AddCertTags(cert.Data.Body.Id, ParseTags(tags))
		app.SaveOrgIndex()
	} else {
		var files []ExportFile
		certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
		files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
		logger.Infof("Export to '%s'", standaloneFile)
		Export(files, standaloneFile)
	}

	return nil
}

func runCSR(args []string) (err error) {
	usage := `
Manages Certificate Signing Requests

Usage:
    pki.io csr [--help]
    pki.io csr new <name> --tags <tags> [--standalone <file>] [--dn-l <locality>] [--dn-st <state>] [--dn-o <org>] [--dn-ou <orgUnit>] [--dn-c <country>] [--dn-street <street>] [--dn-postal <postalCode>]
    pki.io csr list
    pki.io csr show <name> [--export <file>] [--private]
    pki.io csr delete <name> --confirm-delete <reason>
    pki.io csr sign <name> <csrFile> --ca <ca> --tags <tags> [--standalone <file>]

Options:
    --tags <tags>              List of comma-separated tags
    --standalone <file>        Certificate isn't tracked by the Org but is exported to <file>
    --ca <ca>                  Name of CA
    --dn-l <locality>          Locality for DN scope
    --dn-st <state>            State/province for DN scope
    --dn-o <org>               Organization for DN scope
    --dn-ou <orgUnit>          Organizational unit for DN scope
    --dn-c <country>           Country for DN scope
    --dn-street <street>       Street for DN scope
    --dn-postal <postalCode>   Postal code for DN scope
    --confirm-delete <reason>  Reason for deleting node
    --export <file>            Exports cert to <file>
    --private                  Shows private data (e.g. keys)
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["new"].(bool) {
		csrNew(argv)
	} else if argv["list"].(bool) {
		csrList(argv)
	} else if argv["show"].(bool) {
		csrShow(argv)
	} else if argv["delete"].(bool) {
		csrDelete(argv)
	} else if argv["sign"].(bool) {
		csrSign(argv)
	}
	return nil
}
