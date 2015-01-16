package main

import (
	"fmt"
	"github.com/pki-io/pki.io/x509"
)

func csrNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)

	fmt.Println("Creating new CSR")
	csr, _ := x509.NewCSR(nil)
	csr.Data.Body.Name = name
	csr.Generate()

	fmt.Println("Saving local CSR")
	csrContainer, err := org.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		panic(fmt.Sprintf("Could not encrypt CSR: %s", err.Error()))
	}
	if err := fsAPI.StorePrivate(csr.Data.Body.Name, csrContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save CSR: %s", err.Error()))
	}

	fmt.Println("Sending public CSR")
	csrPublic, err := csr.Public()
	if err != nil {
		panic(fmt.Sprintf("Could not get public CSR: %s", err.Error()))
	}

	csrPublicContainer, err := org.SignString(csrPublic.Dump())
	if err != nil {
		panic(fmt.Sprintf("Could not sign public CSR: %s", err.Error()))
	}

	if err := fsAPI.SendPublic(org.Data.Body.Id, csrPublic.Data.Body.Name, csrPublicContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not send public CSR: %s", err.Error()))
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
