package main

import (
	"crypto/x509/pkix"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/core/crypto"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/fs"
	"github.com/pki-io/core/x509"
	"time"
)

func certNew(argv map[string]interface{}) (err error) {
	// TODO - this whole function needs to be refactored
	name := ArgString(argv["<name>"], nil)
	tags := ArgString(argv["--tags"], nil)

	standaloneFile := ArgString(argv["--standalone"], "")
	expiry := ArgInt(argv["--expiry"], 365)

	caName := ArgString(argv["--ca"], "")
	keyType := ArgString(argv["--keytype"], "ec")

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

	cert, err := x509.NewCertificate(nil)
	checkAppFatal("Couldn't create new certificate: %s", err)

	cert.Data.Body.Name = name
	cert.Data.Body.Expiry = expiry

	// TODO - better validation after refactor
	if keyType == "rsa" || keyType == "ec" {
		logger.Infof("Setting key type to %s", keyType)
		cert.Data.Body.KeyType = keyType
	} else {
		checkUserFatal("Invalid key type given. Must be rsa or ec.")
	}

	var files []ExportFile
	certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
	keyFile := fmt.Sprintf("%s-key.pem", cert.Data.Body.Name)
	if caName == "" {
		// Self-signed
		err := cert.Generate(nil, &subject)
		checkAppFatal("Couldn't generate certificate: %s", err)

		// Probably don't need a ca file if self-signed
		// files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
	} else {
		app.LoadOrgIndex()

		caId, err := app.index.org.GetCA(caName)
		checkUserFatal("Couldn't find CA '%s'%.0s", caName, err)

		caContainerJson, err := app.fs.api.GetPrivate(app.entities.org.Data.Body.Id, caId)
		caContainer, err := document.NewContainer(caContainerJson)
		checkAppFatal("Couldn't create container from json: %s", err)

		caJson, err := app.entities.org.VerifyThenDecrypt(caContainer)
		checkAppFatal("Couldn't verify and decrypt ca container: %s", err)

		ca, err := x509.NewCA(caJson)
		checkAppFatal("Couldn't create ca: %s", err)

		err = cert.Generate(ca, &subject)
		checkAppFatal("Couldn't generate certificate: %s", err)
		caFile := fmt.Sprintf("%s-cacert.pem", cert.Data.Body.Name)
		files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(ca.Data.Body.Certificate)})
	}

	files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
	files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})

	if standaloneFile == "" {
		logger.Info("Saving cert")
		certContainer, err := app.entities.org.EncryptThenSignString(cert.Dump(), nil)
		checkAppFatal("Could not encrypt cert: %s", err)

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
		logger.Infof("Export to '%s'", standaloneFile)
		Export(files, standaloneFile)
	}

	return nil
}

func certShow(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	exportFile := ArgString(argv["--export"], "")
	private := ArgBool(argv["--private"], false)

	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()

	certSerial, err := app.index.org.GetCert(name)
	checkUserFatal("Could not find cert: %s%.0s\n", name, err)

	cert := app.GetCert(certSerial)

	if exportFile == "" {
		// TODO - show CA name
		fmt.Printf("Name: %s\n", cert.Data.Body.Name)
		fmt.Printf("ID: %s\n", cert.Data.Body.Id)
		fmt.Printf("Cert expiry period: %d\n", cert.Data.Body.Expiry)
		fmt.Printf("Key type: %s\n", cert.Data.Body.KeyType)
		fmt.Printf("Certficate:\n%s\n", cert.Data.Body.Certificate)

		if private {
			fmt.Printf("Private key:\n%s\n", cert.Data.Body.PrivateKey)
		}
	} else {
		var files []ExportFile
		certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
		keyFile := fmt.Sprintf("%s-key.pem", cert.Data.Body.Name)

		files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})

		if cert.Data.Body.CACertificate != "" {
			caFile := fmt.Sprintf("%s-cacert.pem", cert.Data.Body.Name)
			files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(cert.Data.Body.CACertificate)})
		}

		if private {
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})
		}

		logger.Infof("Export to '%s'", exportFile)
		Export(files, exportFile)
	}

	return nil
}

func certList(argv map[string]interface{}) (err error) {
	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()

	logger.Info("Certs:")
	logger.Flush()
	for name, id := range app.index.org.GetCerts() {
		fmt.Printf("* %s %s\n", name, id)
	}
	return nil
}

func certDelete(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	reason := ArgString(argv["--confirm-delete"], nil)

	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()
	logger.Infof("Deleting cert %s with reason: %s", name, reason)

	certId, err := app.index.org.GetCert(name)
	checkUserFatal("cert %s does not exist%.0s", name, err)

	err = app.fs.api.Authenticate(app.entities.org.Data.Body.Id, "")
	checkAppFatal("Could not authenticate to API as Org: %s", err)

	err = app.fs.api.DeletePrivate(certId)
	checkAppFatal("Could not delete cert: %s", err)

	err = app.index.org.RemoveCert(name)
	checkAppFatal("Could not remove cert: %s", err)
	app.SaveOrgIndex()

	return nil
}

func certImport(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	tags := ArgString(argv["--tags"], nil)

	certFile := ArgString(argv["<certFile>"], nil)
	keyFile := ArgString(argv["<privateKeyFile>"], "")

	logger.Infof("Importing cert %s as %s", certFile, name)
	app := NewAdminApp()
	app.Load()

	// TODO - check if CSR exists for name ....

	cert, _ := x509.NewCertificate(nil)
	cert.Data.Body.Name = name

	ok, err := fs.Exists(certFile)
	checkAppFatal("Could not check file existence for %s: %s", certFile, err)
	if !ok {
		checkUserFatal("File does not exist: %s", certFile)
	}
	certPem, err := fs.ReadFile(certFile)

	importCert, err := x509.PemDecodeX509Certificate([]byte(certPem))
	checkUserFatal("Not a valid certificate PEM for %s: %s", certFile, err)
	// TODO - consider converting cert back to pem to use for consistency

	// We generate a random ID instead of using the serial number because we
	// don't control the serial
	cert.Data.Body.Id = NewID()
	cert.Data.Body.Certificate = certPem
	certExpiry := int(importCert.NotAfter.Sub(importCert.NotBefore) / (time.Hour * 24))
	cert.Data.Body.Expiry = certExpiry

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

		cert.Data.Body.KeyType = string(keyType)
		cert.Data.Body.PrivateKey = keyPem
	}

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

	return nil
}

func runCert(args []string) (err error) {
	usage := `
Manages certificates

Usage: 
    pki.io cert [--help]
    pki.io cert new <name> --tags <tags> [--standalone <file>] [--expiry <days>] [--ca <ca>] [--keytype <type>] [--dn-l <locality>] [--dn-st <state>] [--dn-o <org>] [--dn-ou <orgUnit>] [--dn-c <country>] [--dn-street <street>] [--dn-postal <postalCode>]
    pki.io cert list
    pki.io cert show <name> [--export <file>] [--private]
    pki.io cert delete <name> --confirm-delete <reason>
    pki.io cert import <name> <certFile> [<privateKeyFile>] --tags <tag>

Options:
    --tags <tags>              List of comma-separated tags
    --standalone <file>        Certificate isn't tracked by the Org but is exported to <file>
    --expiry <days>            Expiry period in days [default: 365]
    --ca <ca>                  Name of CA
    --keytype <type>           Key type to use (rsa or ec) [default: ec]
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
		certNew(argv)
	} else if argv["list"].(bool) {
		certList(argv)
	} else if argv["show"].(bool) {
		certShow(argv)
	} else if argv["delete"].(bool) {
		certDelete(argv)
	} else if argv["import"].(bool) {
		certImport(argv)
	}
	return nil
}
