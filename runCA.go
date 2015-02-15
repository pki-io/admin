package main

import (
	"github.com/docopt/docopt-go"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/x509"
	"time"
)

func caNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)
	inTags := argv["--tags"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI()
	admin := LoadAdmin(fsAPI, conf)
	org := LoadOrgPrivate(fsAPI, admin, conf)

	logger.Info("Creating new CA")
	ca, _ := x509.NewCA(nil)
	ca.Data.Body.Name = name
	ca.GenerateRoot(time.Now(), time.Now().AddDate(5, 5, 5))

	logger.Info("Saving CA")
	caContainer, err := org.EncryptThenSignString(ca.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt CA: %s", err))
	}
	if err := fsAPI.SendPrivate(org.Data.Body.Id, ca.Data.Body.Id, caContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save CA: %s", err))
	}

	logger.Info("Updating index")
	indx := LoadOrgIndex(fsAPI, org)
	indx.AddCATags(ca.Data.Body.Id, ParseTags(inTags))
	SaveOrgIndex(fsAPI, org, indx)

	return nil
}

func caSign(argv map[string]interface{}) (err error) {
	caName := argv["<ca>"].(string)
	csrName := argv["<csr>"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI()
	admin := LoadAdmin(fsAPI, conf)
	org := LoadOrgPrivate(fsAPI, admin, conf)

	logger.Info("Getting CA")
	caJson, err := fsAPI.GetPrivate(org.Data.Body.Id, caName)
	if err != nil {
		panic(logger.Errorf("Could not get encrypted CA: %s", err))
	}

	caContainer, err := document.NewContainer(caJson)
	if err != nil {
		panic(logger.Errorf("Could not create CA container: %s", err))
	}

	if err := org.Verify(caContainer); err != nil {
		panic(logger.Errorf("Could not verify CA: %s", err))
	}

	decryptedCAJson, err := org.Decrypt(caContainer)
	if err != nil {
		panic(logger.Errorf("Could not decrypt CA: %s", err))
	}

	ca, err := x509.NewCA(decryptedCAJson)
	if err != nil {
		panic(logger.Errorf("Could not create CA from JSON: %s", err))
	}

	logger.Info("Getting CSR")
	csrJson, err := fsAPI.GetPublic(org.Data.Body.Id, csrName)
	if err != nil {
		panic(logger.Errorf("Could not get CSR: %s", err))
	}

	csrContainer, err := document.NewContainer(csrJson)
	if err := org.Verify(csrContainer); err != nil {
		panic(logger.Errorf("Could not verify CSR: %s", err))
	}

	csr, err := x509.NewCSR(csrContainer.Data.Body)
	if err != nil {
		panic(logger.Errorf("Could not create CSR from JSON: %s", err))
	}

	logger.Info("Signing CSR")
	cert, err := ca.Sign(csr)
	if err != nil {
		panic(logger.Errorf("Could not sign CSR: %s", err))
	}

	certContainer, err := org.SignString(cert.Dump())
	if err != nil {
		panic(logger.Errorf("Could not sign cert: %s", err))
	}

	logger.Info("Saving certificate")
	if err := fsAPI.SendPrivate(admin.Data.Body.Id, "cert_"+csr.Data.Body.Name, certContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save cert: %s", err))
	}
	return nil
}

// CA related commands
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
	} else if argv["sign"].(bool) {
		caSign(argv)
	}
	return nil
}
