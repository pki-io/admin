package main

import (
	"fmt"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/x509"
)

func certShow(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPublic(fsAPI, admin)

	certJson, err := fsAPI.LoadPrivate(name)
	if err != nil {
		panic(fmt.Sprintf("Could not load cert json: %s", err.Error()))
	}

	certContainer, err := document.NewContainer(certJson)
	if err != nil {
		panic(fmt.Sprintf("Could not create cert container: %s", err.Error()))
	}

	if err := org.Verify(certContainer); err != nil {
		panic(fmt.Sprintf("Could not verify container: %s", err.Error()))
	}

	cert, err := x509.NewCertificate(certContainer.Data.Body)
	if err != nil {
		panic(fmt.Sprintf("Could not create cert: %s", err.Error()))
	}

	fmt.Printf("Name: %s\n", cert.Data.Body.Name)
	fmt.Printf("Id: %s\n", cert.Data.Body.Id)
	fmt.Printf("Certificate:\n%s\n", cert.Data.Body.Certificate)

	return nil
}

func runCert(argv map[string]interface{}) (err error) {
	if argv["show"].(bool) {
		certShow(argv)
	}
	return nil
}
