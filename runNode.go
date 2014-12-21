package main

import (
	"fmt"
	"pki.io/entity"
	"pki.io/fs"
	"pki.io/x509"
)

func nodeNewCSR(fsAPI *fs.FsAPI, node, org *entity.Entity) {
	fmt.Println("Creating new CSR")
	csr, err := x509.NewCSR(nil)
	if err != nil {
		panic(fmt.Sprintf("Could not generate CSR: %s", err.Error()))
	}

	csr.Data.Body.Name = node.Data.Body.Name
	csr.Generate()

	fmt.Println("Saving local CSR")
	csrContainer, err := node.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		panic(fmt.Sprintf("Could not encrypt CSR: %s", err.Error()))
	}
	if err := fsAPI.SendPrivate(node.Data.Body.Id, csr.Data.Body.Name, csrContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save CSR: %s", err.Error()))
	}

	fmt.Println("Sending public CSR")
	csrPublic, err := csr.Public()
	if err != nil {
		panic(fmt.Sprintf("Could not get public CSR: %s", err.Error()))
	}

	csrPublicContainer, err := node.SignString(csrPublic.Dump())
	if err != nil {
		panic(fmt.Sprintf("Could not sign public CSR: %s", err.Error()))
	}

	if err := fsAPI.SendPublic(org.Data.Body.Id, csrPublic.Data.Body.Name, csrPublicContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not send public CSR: %s", err.Error()))
	}
}
func nodeGenerateCSRs(fsAPI *fs.FsAPI, node, org *entity.Entity) error {
	nodeNewCSR(fsAPI, node, org)
	return nil
}

func nodeNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)

	fmt.Println("Creating new node")
	node, err := entity.New(nil)
	if err != nil {
		panic(fmt.Sprintf("Could not create node entity: %s:", err.Error()))
	}
	node.Data.Body.Name = name
	node.Data.Body.Id = NewID()

	fmt.Println("Generating node keys")
	if err := node.GenerateKeys(); err != nil {
		panic(fmt.Sprintf("Could not generate node keys: %s", err.Error()))
	}

	fmt.Println("Saving node config")
	conf.AddNode(node.Data.Body.Name, node.Data.Body.Id)
	if err := conf.Save(); err != nil {
		panic(fmt.Sprintf("Could not save admin config: %s", err.Error()))
	}

	fmt.Println("Creating public registration document")
	publicNode, err := node.Public()
	if err != nil {
		panic(fmt.Sprintf("Could get public node: %s", err.Error()))
	}

	fmt.Println("Sending public document")
	if err := fsAPI.SendPublic(org.Data.Body.Id, node.Data.Body.Name, publicNode.Dump()); err != nil {
		panic(fmt.Sprintf("Could not send document to org: %s", err.Error()))
	}
	fmt.Println("Saving node")
	if err := fsAPI.WriteLocal(node.Data.Body.Name, node.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save node: %s", err.Error()))
	}

	// create crs
	nodeGenerateCSRs(fsAPI, node, org)

	return nil
}

// Node related commands
func runNode(argv map[string]interface{}) (err error) {
	if argv["new"].(bool) {
		nodeNew(argv)
	}
	return nil
}
