package main

import (
	"github.com/pki-io/pki.io/x509"
)

func csrNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)

	logger.Info("Creating new CSR")
	csr, _ := x509.NewCSR(nil)
	csr.Data.Body.Name = name
	csr.Generate()

	logger.Info("Saving local CSR")
	csrContainer, err := org.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt CSR: %s", err))
	}
	if err := fsAPI.StorePrivate(csr.Data.Body.Name, csrContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save CSR: %s", err))
	}

	logger.Info("Sending public CSR")
	csrPublic, err := csr.Public()
	if err != nil {
		panic(logger.Errorf("Could not get public CSR: %s", err))
	}

	csrPublicContainer, err := org.SignString(csrPublic.Dump())
	if err != nil {
		panic(logger.Errorf("Could not sign public CSR: %s", err))
	}

	if err := fsAPI.SendPublic(org.Data.Body.Id, csrPublic.Data.Body.Name, csrPublicContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not send public CSR: %s", err))
	}

	return nil
}

// CSR related commands
func runCSR(argv map[string]interface{}) (err error) {
	if argv["new"].(bool) {
		csrNew(argv)
	}
	return nil
}
