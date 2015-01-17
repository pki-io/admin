package main

import (
	"fmt"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/x509"
	"time"
)

func caNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)
	inTags := argv["--tags"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)

	fmt.Println("Creating new CA")
	ca, _ := x509.NewCA(nil)
	ca.Data.Body.Name = name
	ca.GenerateRoot(time.Now(), time.Now().AddDate(5, 5, 5))

	fmt.Println("Saving CA")
	caContainer, err := org.EncryptThenSignString(ca.Dump(), nil)
	if err != nil {
		panic(fmt.Sprintf("Could not encrypt CA: %s", err.Error()))
	}
	if err := fsAPI.SendPrivate(org.Data.Body.Id, ca.Data.Body.Id, caContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save CA: %s", err.Error()))
	}

	fmt.Println("Updating index")
	indx := LoadIndex(fsAPI, org)
	indx.AddCATags(ca.Data.Body.Id, ParseTags(inTags))
	SaveIndex(fsAPI, org, indx)

	return nil
}

func caSign(argv map[string]interface{}) (err error) {
	caName := argv["<ca>"].(string)
	csrName := argv["<csr>"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)

	fmt.Println("Getting CA")
	caJson, err := fsAPI.GetPrivate(org.Data.Body.Id, caName)
	if err != nil {
		panic(fmt.Sprintf("Could not get encrypted CA: %s", err.Error()))
	}

	caContainer, err := document.NewContainer(caJson)
	if err != nil {
		panic(fmt.Sprintf("Could not create CA container: %s", err.Error()))
	}

	if err := org.Verify(caContainer); err != nil {
		panic(fmt.Sprintf("Could not verify CA: %s", err.Error()))
	}

	decryptedCAJson, err := org.Decrypt(caContainer)
	if err != nil {
		panic(fmt.Sprintf("Could not decrypt CA: %s", err.Error()))
	}

	ca, err := x509.NewCA(decryptedCAJson)
	if err != nil {
		panic(fmt.Sprintf("Could not create CA from JSON: %s", err.Error()))
	}

	fmt.Println("Getting CSR")
	csrJson, err := fsAPI.GetPublic(org.Data.Body.Id, csrName)
	if err != nil {
		panic(fmt.Sprintf("Could not get CSR: %s", err.Error()))
	}

	csrContainer, err := document.NewContainer(csrJson)
	if err := org.Verify(csrContainer); err != nil {
		panic(fmt.Sprintf("Could not verify CSR: %s", err.Error()))
	}

	csr, err := x509.NewCSR(csrContainer.Data.Body)
	if err != nil {
		panic(fmt.Sprintf("Could not create CSR from JSON: %s", err.Error()))
	}

	fmt.Println("Signing CSR")
	cert, err := ca.Sign(csr)
	if err != nil {
		panic(fmt.Sprintf("Could not sign CSR: %s", err.Error()))
	}

	certContainer, err := org.SignString(cert.Dump())
	if err != nil {
		panic(fmt.Sprintf("Could not sign cert: %s", err.Error()))
	}

	fmt.Println("Saving certificate")
	if err := fsAPI.SendPrivate(admin.Data.Body.Id, "cert_"+csr.Data.Body.Name, certContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save cert: %s", err.Error()))
	}
	return nil
}

// CA related commands
func runCA(argv map[string]interface{}) (err error) {
	if argv["new"].(bool) {
		caNew(argv)
	} else if argv["sign"].(bool) {
		caSign(argv)
	}
	return nil
}
