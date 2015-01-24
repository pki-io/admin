package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/entity"
	"github.com/pki-io/pki.io/fs"
	n "github.com/pki-io/pki.io/node"
	"github.com/pki-io/pki.io/x509"
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

func nodeProcessRegistrations(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPublic(fsAPI, admin)

	// Need to rename the privte key files a bit. Then look them up via the config. Would do a search until
	// name is matchedc
	//nodeId := conf.Data.Nodes[0].Id // Shouldn't hardcode, assuming one node for now

	nodeJson, err := fsAPI.ReadLocal(name)
	if err != nil {
		panic(fmt.Sprintf("Could not read node file: %s", err.Error()))
	}

	node, err := n.New(nodeJson)
	if err != nil {
		panic(fmt.Sprintf("Could not load node json: %s", err.Error()))
	}

	fsAPI.Id = node.Data.Body.Id

	for {
		size, err := fsAPI.IncomingSize("registrations")
		if err != nil {
			panic(fmt.Sprintf("Can't get queue size: %s", err.Error()))
		}

		fmt.Printf("Found %d registrations to process\n", size)
		if size > 0 {
			nodeContainerJson, err := fsAPI.PopIncoming("registrations")
			if err != nil {
				panic(fmt.Sprintf("Can't pop registrations: %s", err.Error()))
			}

			nodeContainer, err := document.NewContainer(nodeContainerJson)
			if err != nil {
				panic(fmt.Sprintf("Can't create node container: %s", err.Error()))
			}

			if err := org.Verify(nodeContainer); err != nil {
				panic(fmt.Sprintf("Can't verify node registration: %s", err.Error()))
			}

			nodeReg, err := n.New(nodeContainer.Data.Body)
			if err != nil {
				panic(fmt.Sprintf("Can't node registration: %s", err.Error()))
			}
			fmt.Println(nodeReg)

			/*node.Data.Body.Tags = nodeReg.Data.Body.Tags

			fmt.Println("Saving node")
			if err := fsAPI.WriteLocal(node.Data.Body.Name, node.Dump()); err != nil {
				panic(fmt.Sprintf("Could not save node: %s", err.Error()))
			}*/
		}
	}

	return nil
}

func nodeProcessCerts(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPublic(fsAPI, admin)

	// Need to rename the privte key files a bit. Then look them up via the config. Would do a search until
	// name is matchedc
	//nodeId := conf.Data.Nodes[0].Id // Shouldn't hardcode, assuming one node for now

	nodeJson, err := fsAPI.ReadLocal(name)
	if err != nil {
		panic(fmt.Sprintf("Could not read node file: %s", err.Error()))
	}

	node, err := n.New(nodeJson)
	if err != nil {
		panic(fmt.Sprintf("Could not load node json: %s", err.Error()))
	}

	fsAPI.Id = node.Data.Body.Id

	// For each income cert
	for {
		size, err := fsAPI.IncomingSize("certs")
		if err != nil {
			panic(fmt.Sprintf("Can't get queue size: %s", err.Error()))
		}

		fmt.Printf("Found %d certs to process\n", size)
		if size > 0 {
			certContainerJson, err := fsAPI.PopIncoming("certs")
			if err != nil {
				panic(fmt.Sprintf("Can't pop cert: %s", err.Error()))
			}

			certContainer, err := document.NewContainer(certContainerJson)
			if err != nil {
				panic(fmt.Sprintf("Can't create cert container: %s", err.Error()))
			}
			if err := org.Verify(certContainer); err != nil {
				panic(fmt.Sprintf("Cert didn't verify: %s", err.Error()))
			}

			cert, err := x509.NewCertificate(certContainer.Data.Body)
			if err != nil {
				panic(fmt.Sprintf("Can't load certificate: %s", err.Error()))
			}

			// Read local CSR to get private key
			csrContainerJson, err := fsAPI.GetPrivate(fsAPI.Id, cert.Data.Body.Id)
			if err != nil {
				panic(fmt.Sprintf("Couldn't load csr file: %s", err.Error()))
			}

			csrContainer, err := document.NewContainer(csrContainerJson)
			if err != nil {
				panic(fmt.Sprintf("Couldn't load CSR container: %s", err.Error()))
			}

			csrJson, err := node.VerifyThenDecrypt(csrContainer)
			if err != nil {
				panic(fmt.Sprintf("Couldn't verify and decrypt container: %s", err.Error()))
			}

			csr, err := x509.NewCSR(csrJson)
			if err != nil {
				panic(fmt.Sprintf("Couldn't load csr: %s", err.Error()))
			}

			// Set the private key from the csr
			cert.Data.Body.PrivateKey = csr.Data.Body.PrivateKey

			// Reuse container
			updatedCertContainer, err := node.EncryptThenSignString(cert.Dump(), nil)
			if err != nil {
				panic(fmt.Sprintf("Could not encrypt then sign cert: %s", err.Error()))
			}

			if err := fsAPI.SendPrivate(node.Data.Body.Id, cert.Data.Body.Id, updatedCertContainer.Dump()); err != nil {
				panic(fmt.Sprintf("Could save cert: %s", err.Error()))
			}

		} else {
			break
		}
	}
	fmt.Println("Creating CSRs")
	nodeGenerateCSRs(fsAPI, node, org)
	return nil
}

func nodeShow(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)
	certId := argv["--cert"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	//admin := LoadAdmin(fsAPI)
	//org := LoadOrgPublic(fsAPI, admin)

	// Need to rename the privte key files a bit. Then look them up via the config. Would do a search until
	// name is matchedc
	//nodeId := conf.Data.Nodes[0].Id // Shouldn't hardcode, assuming one node for now

	nodeJson, err := fsAPI.ReadLocal(name)
	if err != nil {
		panic(fmt.Sprintf("Could not read node file: %s", err.Error()))
	}

	node, err := n.New(nodeJson)
	if err != nil {
		panic(fmt.Sprintf("Could not load node json: %s", err.Error()))
	}

	fsAPI.Id = node.Data.Body.Id

	certContainerJson, err := fsAPI.GetPrivate(fsAPI.Id, certId)
	if err != nil {
		panic(fmt.Sprintf("Could not load cert container json: %s", err.Error()))
	}

	certContainer, err := document.NewContainer(certContainerJson)
	if err != nil {
		panic(fmt.Sprintf("Could not load cert container: %s", err.Error()))
	}

	certJson, err := node.VerifyThenDecrypt(certContainer)
	if err != nil {
		panic(fmt.Sprintf("Could not verify and decrypt: %s", err.Error()))
	}

	cert, err := x509.NewCertificate(certJson)
	if err != nil {
		panic(fmt.Sprintf("Could not load cert: %s", err.Error()))
	}

	fmt.Printf("Certificate:\n%s\n\n", cert.Data.Body.Certificate)
	fmt.Printf("Private Key:\n%s\n\n", cert.Data.Body.PrivateKey)

	return nil
}

// Node related commands
func runNode(args []string) (err error) {

	usage := `
Usage:
    pki.io node [--help]
    pki.io node new <name> --pairing-id=<id> --pairing-key=<key>
    pki.io node run
    pki.io node show --name=<name> --cert=<id>

Options:
    --pairing-id=<id>   Pairing ID
    --pairing-key=<key> Pairing Key
    --name=<name>       Node name
    --cert=<cert>       Certificate ID
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["new"].(bool) {
		nodeNew(argv)
	} else if argv["run"].(bool) {
		nodeProcessCerts(argv)
	} else if argv["show"].(bool) {
		nodeShow(argv)
	}
	return nil
}
