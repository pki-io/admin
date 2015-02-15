package main

import (
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/x509"
)

func certShow(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI()
	admin := LoadAdmin(fsAPI, conf)
	org := LoadOrgPublic(fsAPI, admin, conf)

	certJson, err := fsAPI.LoadPrivate(name)
	if err != nil {
		panic(logger.Errorf("Could not load cert json: %s", err))
	}

	certContainer, err := document.NewContainer(certJson)
	if err != nil {
		panic(logger.Errorf("Could not create cert container: %s", err))
	}

	if err := org.Verify(certContainer); err != nil {
		panic(logger.Errorf("Could not verify container: %s", err))
	}

	cert, err := x509.NewCertificate(certContainer.Data.Body)
	if err != nil {
		panic(logger.Errorf("Could not create cert: %s", err))
	}

	logger.Infof("Name: %s\n", cert.Data.Body.Name)
	logger.Infof("Id: %s\n", cert.Data.Body.Id)
	logger.Infof("Certificate:\n%s\n", cert.Data.Body.Certificate)

	return nil
}

func runCert(argv map[string]interface{}) (err error) {
	if argv["show"].(bool) {
		certShow(argv)
	}
	return nil
}
