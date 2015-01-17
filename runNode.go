package main

import (
	"fmt"
	"pki.io/entity"
	"pki.io/fs"
	n "pki.io/node"
	"pki.io/x509"
)

const MinCSRs = 5

func nodeNewCSR(fsAPI *fs.FsAPI, node, org *entity.Entity) {
	fmt.Println("Creating new CSR")
	csr, err := x509.NewCSR(nil)
	if err != nil {
		panic(fmt.Sprintf("Could not generate CSR: %s", err.Error()))
	}

	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = node.Data.Body.Name
	csr.Generate()

	fmt.Println("Saving local CSR")
	csrContainer, err := node.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		panic(fmt.Sprintf("Could not encrypt CSR: %s", err.Error()))
	}
	if err := fsAPI.SendPrivate(node.Data.Body.Id, csr.Data.Body.Id, csrContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save CSR: %s", err.Error()))
	}

	fmt.Println("Pushing public CSR")
	csrPublic, err := csr.Public()
	if err != nil {
		panic(fmt.Sprintf("Could not get public CSR: %s", err.Error()))
	}

	csrPublicContainer, err := node.SignString(csrPublic.Dump())
	if err != nil {
		panic(fmt.Sprintf("Could not sign public CSR: %s", err.Error()))
	}

	if err := fsAPI.PushOutgoing("csrs", csrPublicContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not send public CSR: %s", err.Error()))
	}
}
func nodeGenerateCSRs(fsAPI *fs.FsAPI, node, org *entity.Entity) error {
	numCSRs, err := fsAPI.OutgoingSize(fsAPI.Id, "csrs")
	if err != nil {
		panic(fmt.Sprintf("Could not get csr queue size: %s", err.Error()))
	}

	for i := 0; i < MinCSRs-numCSRs; i++ {
		nodeNewCSR(fsAPI, node, org)
	}
	return nil
}

func nodeNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)
	pairingId := argv["--pairing-id"].(string)
	pairingKey := argv["--pairing-key"].(string)
	//inTags := argv["--tags"].(string)

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

	// Acting on behalf of the node now, so all API is done for the node
	//adminId := fsAPI.Id
	fsAPI.Id = node.Data.Body.Id

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
	reg, err := n.NewRegistration(node)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create registration: %s", err.Error()))
	}
	if err := reg.Authenticate(pairingId, pairingKey); err != nil {
		panic(fmt.Sprintf("Couldn't authenticate registration: %s", err.Error()))
	}

	fmt.Println("Pushing public document to org")
	if err := fsAPI.PushIncoming(org.Data.Body.Id, "registration", reg.Dump()); err != nil {
		panic(fmt.Sprintf("Could not push document to org: %s", err.Error()))
	}
	fmt.Println("Saving node")
	if err := fsAPI.WriteLocal(node.Data.Body.Name, node.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save node: %s", err.Error()))
	}

	// create crs
	fmt.Println("Creating CSRs")
	nodeGenerateCSRs(fsAPI, node, org)

	return nil
}

func nodeInstallCerts(argv map[string]interface{}) (err error) {
	/*name := argv["--name"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPublic(fsAPI, admin)*/

	// Read node file from name
	// For each income cert
	// verify container
	// write cert to private store
	return nil
}

// Node related commands
func runNode(argv map[string]interface{}) (err error) {
	if argv["new"].(bool) {
		nodeNew(argv)
	} else if argv["install-certs"].(bool) {
		nodeInstallCerts(argv)
	}
	return nil
}
