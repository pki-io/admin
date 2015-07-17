package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/core/crypto"
	"github.com/pki-io/core/fs"
	"github.com/pki-io/core/x509"
	"time"
)

func caNew(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	tags := ArgString(argv["--tags"], nil)

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

	err = app.fs.api.Authenticate(app.entities.org.Data.Body.Id, "")
	checkAppFatal("Could not authenticate to API as Org: %s", err)

	err = app.fs.api.StorePrivate(ca.Data.Body.Id, caContainer.Dump())
	checkAppFatal("Could not save CA: %s", err)

	logger.Info("Updating index")
	app.LoadOrgIndex()
	app.index.org.AddCA(ca.Data.Body.Name, ca.Data.Body.Id)
	app.index.org.AddCATags(ca.Data.Body.Id, ParseTags(tags))
	app.SaveOrgIndex()

	return nil
}

func caList(argv map[string]interface{}) (err error) {
	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()

	logger.Info("CAs:")
	logger.Flush()
	for name, id := range app.index.org.GetCAs() {
		fmt.Printf("* %s %s\n", name, id)
	}
	return nil
}

func caShow(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	exportFile := ArgString(argv["--export"], "")
	private := ArgBool(argv["--private"], false)

	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()

	// TODO - refactor
	caSerial, err := app.index.org.GetCA(name)
	checkUserFatal("Could not find CA: %s%.0s\n", name, err)

	ca := app.GetCA(caSerial)

	if exportFile == "" {
		fmt.Printf("Name: %s\n", ca.Data.Body.Name)
		fmt.Printf("ID: %s\n", ca.Data.Body.Id)
		fmt.Printf("CA expiry period: %d\n", ca.Data.Body.CAExpiry)
		fmt.Printf("Cert expiry period: %d\n", ca.Data.Body.CertExpiry)
		fmt.Printf("Key type: %s\n", ca.Data.Body.KeyType)
		fmt.Printf("DN country: %s\n", ca.Data.Body.DNScope.Country)
		fmt.Printf("DN organization: %s\n", ca.Data.Body.DNScope.Organization)
		fmt.Printf("DN organizational unit: %s\n", ca.Data.Body.DNScope.OrganizationalUnit)
		fmt.Printf("DN locality: %s\n", ca.Data.Body.DNScope.Locality)
		fmt.Printf("DN province: %s\n", ca.Data.Body.DNScope.Province)
		fmt.Printf("DN street address: %s\n", ca.Data.Body.DNScope.StreetAddress)
		fmt.Printf("DN postal code: %s\n", ca.Data.Body.DNScope.PostalCode)
		fmt.Printf("Certficate:\n%s\n", ca.Data.Body.Certificate)

		if private {
			fmt.Printf("Private key:\n%s\n", ca.Data.Body.PrivateKey)
		}
	} else {
		var files []ExportFile
		certFile := fmt.Sprintf("%s-cert.pem", ca.Data.Body.Name)
		keyFile := fmt.Sprintf("%s-key.pem", ca.Data.Body.Name)

		files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(ca.Data.Body.Certificate)})

		if private {
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(ca.Data.Body.PrivateKey)})
		}

		logger.Infof("Export to '%s'", exportFile)
		Export(files, exportFile)

	}

	return nil
}

func caDelete(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	reason := ArgString(argv["--confirm-delete"], nil)

	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()
	logger.Infof("Deleting CA %s with reason: %s", name, reason)
	logger.Info("Note: This does not revoke existing certificates signed by the CA")

	caId, err := app.index.org.GetCA(name)
	checkUserFatal("CA %s does not exist%.0s", name, err)

	err = app.fs.api.Authenticate(app.entities.org.Data.Body.Id, "")
	checkAppFatal("Could not authenticate to API as Org: %s", err)

	err = app.fs.api.DeletePrivate(caId)
	checkAppFatal("Could not delete CA: %s", err)

	err = app.index.org.RemoveCA(name)
	checkAppFatal("Could not remove CA: %s", err)
	app.SaveOrgIndex()

	return nil
}

func caImport(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	tags := ArgString(argv["--tags"], nil)

	certFile := ArgString(argv["<certFile>"], nil)
	keyFile := ArgString(argv["<privateKeyFile>"], "")

	certExpiry := ArgInt(argv["--cert-expiry"], 90)

	dnLocality := ArgString(argv["--dn-l"], "")
	dnState := ArgString(argv["--dn-st"], "")
	dnOrg := ArgString(argv["--dn-o"], "")
	dnOrgUnit := ArgString(argv["--dn-ou"], "")
	dnCountry := ArgString(argv["--dn-c"], "")
	dnStreet := ArgString(argv["--dn-street"], "")
	dnPostal := ArgString(argv["--dn-postal"], "")

	logger.Infof("Importing CA %s as %s", certFile, name)
	app := NewAdminApp()
	app.Load()

	ca, _ := x509.NewCA(nil)
	ca.Data.Body.Name = name

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

	ok, err := fs.Exists(certFile)
	checkAppFatal("Could not check file existence for %s: %s", certFile, err)
	if !ok {
		checkUserFatal("File does not exist: %s", certFile)
	}
	certPem, err := fs.ReadFile(certFile)

	cert, err := x509.PemDecodeX509Certificate([]byte(certPem))
	checkUserFatal("Not a valid certificate PEM for %s: %s", certFile, err)
	// TODO - consider converting cert back to pem to use for consistency

	// We generate a random ID instead of using the serial number because we
	// don't control the serial
	ca.Data.Body.Id = NewID()
	ca.Data.Body.Certificate = certPem
	ca.Data.Body.CertExpiry = certExpiry
	caExpiry := int(cert.NotAfter.Sub(cert.NotBefore) / (time.Hour * 24))
	ca.Data.Body.CAExpiry = caExpiry

	if keyFile != "" {
		ok, err = fs.Exists(keyFile)
		checkAppFatal("Could not check file existence for %s: %s", keyFile, err)
		if !ok {
			checkUserFatal("File does not exist: %s", keyFile)
		}
		keyPem, err := fs.ReadFile(keyFile)

		key, err := crypto.PemDecodePrivate([]byte(keyPem))
		checkUserFatal("Not a valid private key PEM for %s: %s", keyFile, err)
		// TODO - consider converting key back to pem to use for consistency

		keyType, err := crypto.GetKeyType(key)
		checkUserFatal("Unknow private key file for %s: %s", keyFile, err)

		ca.Data.Body.KeyType = string(keyType)
		ca.Data.Body.PrivateKey = keyPem
	}

	logger.Info("Saving CA")
	caContainer, err := app.entities.org.EncryptThenSignString(ca.Dump(), nil)
	checkAppFatal("Could not encrypt CA: %s", err)

	err = app.fs.api.Authenticate(app.entities.org.Data.Body.Id, "")
	checkAppFatal("Could not authenticate to API as Org: %s", err)

	err = app.fs.api.StorePrivate(ca.Data.Body.Id, caContainer.Dump())
	checkAppFatal("Could not save CA: %s", err)

	logger.Info("Updating index")
	app.LoadOrgIndex()
	app.index.org.AddCA(ca.Data.Body.Name, ca.Data.Body.Id)
	app.index.org.AddCATags(ca.Data.Body.Id, ParseTags(tags))
	app.SaveOrgIndex()

	return nil
}

func runCA(args []string) (err error) {
	usage := `
Manages Certificate Authorities

Usage: 
    pki.io ca [--help]
    pki.io ca new <name> --tags <tags> [--ca-expiry <days>] [--cert-expiry <days>] [--dn-l <locality>] [--dn-st <state>] [--dn-o <org>] [--dn-ou <orgUnit>] [--dn-c <country>] [--dn-street <street>] [--dn-postal <postalCode>]
    pki.io ca list
    pki.io ca show <name> [--export <file>] [--private]
    pki.io ca delete <name> --confirm-delete <reason>
    pki.io ca import <name> <certFile> [<privateKeyFile>] --tags <tags> [--ca-expiry <days>] [--cert-expiry <days>] [--dn-l <locality>] [--dn-st <state>] [--dn-o <org>] [--dn-ou <orgUnit>] [--dn-c <country>] [--dn-street <street>] [--dn-postal <postalCode>]

Options:
    --tags <tags>              List of comma-separated tags
    --ca-expiry <days>         Expiry period for CA in days [default: 365]
    --cert-expiry <days>       Expiry period for certs in day [default: 90]
    --dn-l <locality>          Locality for DN scope
    --dn-st <state>            State/province for DN scope
    --dn-o <org>               Organization for DN scope
    --dn-ou <orgUnit>          Organizational unit for DN scope
    --dn-c <country>           Country for DN scope
    --dn-street <street>       Street for DN scope
    --dn-postal <postalCode>   Postal code for DN scope
    --confirm-delete <reason>  Reason for deleting node
    --export <file>            Exports cert to <file>
    --private                  Show private data (e.g. keys)
`
	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["new"].(bool) {
		caNew(argv)
	} else if argv["list"].(bool) {
		caList(argv)
	} else if argv["show"].(bool) {
		caShow(argv)
	} else if argv["delete"].(bool) {
		caDelete(argv)
	} else if argv["import"].(bool) {
		caImport(argv)
	}
	return nil
}
