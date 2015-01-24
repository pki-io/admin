package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/node"
	"github.com/pki-io/pki.io/x509"
)

func orgShow(argv map[string]interface{}) (err error) {

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)
	index := LoadIndex(fsAPI, org)

	fmt.Printf("Name: %s\n", org.Data.Body.Name)
	fmt.Printf("Id: %s\n", org.Data.Body.Id)
	fmt.Printf("Public Signing Key:\n%s\n", org.Data.Body.PublicSigningKey)
	fmt.Printf("Public Encryption Key:\n%s\n", org.Data.Body.PublicEncryptionKey)

	fmt.Println("Tags for CAs:")
	for k, v := range index.Data.Body.Tags.CAForward {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("CA tags:")
	for k, v := range index.Data.Body.Tags.CAReverse {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("Tags for entities:")
	for k, v := range index.Data.Body.Tags.EntityForward {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("Entity tags:")
	for k, v := range index.Data.Body.Tags.EntityReverse {
		fmt.Printf("  %s => %s\n", k, v)
	}

	fmt.Println("Pairing keys:")
	for k, v := range index.Data.Body.PairingKeys {
		fmt.Printf("  %s => %s\n", k, v)
	}
	return nil
}

func orgRegisterNodes(argv map[string]interface{}) (err error) {
	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)
	indx := LoadIndex(fsAPI, org)

	fsAPI.Id = org.Data.Body.Id

	fmt.Println("Registering nodes")
	for {
		size, err := fsAPI.IncomingSize("registration")
		if err != nil {
			panic(fmt.Sprintf("Can't get queue size: %s", err.Error()))
		}
		fmt.Printf("Found %d nodes to register\n", size)
		if size > 0 {
			regJson, err := fsAPI.PopIncoming("registration")
			if err != nil {
				panic(fmt.Sprintf("Can't pop registration: %s", err.Error()))
			}

			nodeReg, err := node.NewRegistration(regJson)
			if err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(fmt.Sprintf("Can't load registration: %s", err.Error()))
			}

			pairingId := nodeReg.Data.Options.PairingId
			pairingKey := indx.Data.Body.PairingKeys[pairingId]
			if err := nodeReg.Verify(pairingKey.Key); err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(fmt.Sprintf("Couldn't verify registration: %s", err.Error()))
			}

			node, err := node.NewFromRegistration(nodeReg)
			if err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(fmt.Sprintf("Couldn't create node from registration: %s", err.Error()))
			}
			//node.Data.Body.Tags = pairingKey.Tags
			indx.AddEntityTags(node.Data.Body.Id, pairingKey.Tags)
			// Create node container
			// Sign node container
			// Push signed container to node's incoming queue

			nodeContainer, err := document.NewContainer(nil)
			if err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(fmt.Sprintf("Couldn't create node container: %s", err.Error()))
			}

			nodeContainer.Data.Body = node.Dump()
			if err := org.Sign(nodeContainer); err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(fmt.Sprintf("Couldn't sign node container: %s", err.Error()))

			}
			/*if err := fsAPI.PushIncoming(node.Data.Body.Id, "registration"); err != nil {
				panic(fmt.Sprintf("Couldn't push node: %s", err.Error()))
			}*/

			// For each tag, look for CAs
			for _, tag := range pairingKey.Tags {
				for _, caId := range indx.Data.Body.Tags.CAForward[tag] {

					// For each CA get a CSR for node
					csrContainerJson, err := fsAPI.PopOutgoing(node.Data.Body.Id, "csrs")
					if err != nil {
						panic(fmt.Sprintf("Couldn't get a csr: %s", err.Error()))
					}

					csrContainer, err := document.NewContainer(csrContainerJson)
					if err != nil {
						panic(fmt.Sprintf("Couldn't create container from json: %s", err.Error()))
					}

					if err := node.Verify(csrContainer); err != nil {
						panic(fmt.Sprintf("Couldn't verify CSR: %s", err.Error()))
					}

					csrJson := csrContainer.Data.Body
					csr, err := x509.NewCSR(csrJson)
					if err != nil {
						panic(fmt.Sprintf("Couldn't create csr from json: %s", err.Error()))
					}

					// Get the CA
					caContainerJson, err := fsAPI.GetPrivate(fsAPI.Id, caId)
					caContainer, err := document.NewContainer(caContainerJson)
					if err != nil {
						panic(fmt.Sprintf("Couldn't create container from json: %s", err.Error()))
					}
					caJson, err := org.VerifyThenDecrypt(caContainer)
					if err != nil {
						panic(fmt.Sprintf("Couldn't verify and decrypt ca container: %s", err.Error()))
					}

					ca, err := x509.NewCA(caJson)
					if err != nil {
						panic(fmt.Sprintf("Couldn't create ca: %s", err.Error()))
					}

					// Create a cert
					cert, err := ca.Sign(csr)
					if err != nil {
						panic(fmt.Sprintf("Couldn't sign csr: %s", err.Error()))
					}

					// Sign cert
					certContainer, err := document.NewContainer(nil)
					if err != nil {
						panic(fmt.Sprintf("Couldn't create cert container: %s", err.Error()))
					}
					certContainer.Data.Options.Source = org.Data.Body.Id
					certContainer.Data.Body = cert.Dump()
					if err := org.Sign(certContainer); err != nil {
						panic(fmt.Sprintf("Couldn't sign cert container: %s", err.Error()))
					}

					// Push cert to node's incoming queue
					if err := fsAPI.PushIncoming(node.Data.Body.Id, "certs", certContainer.Dump()); err != nil {
						panic(fmt.Sprintf("Couldn't push cert to node: %s", err.Error()))

					}
				}
			}
		} else {
			break
		}
	}

	SaveIndex(fsAPI, org, indx)
	return nil
}

func runOrg(args []string) (err error) {
	usage := `
Usage:
    pki.io org [--help]
    pki.io org run
    pki.io org show
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["show"].(bool) {
		orgShow(argv)
	} else if argv["run"].(bool) {
		orgRegisterNodes(argv)
	}
	return nil
}
