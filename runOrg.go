package main

import (
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

	logger.Infof("Name: %s\n", org.Data.Body.Name)
	logger.Infof("Id: %s\n", org.Data.Body.Id)
	logger.Infof("Public Signing Key:\n%s\n", org.Data.Body.PublicSigningKey)
	logger.Infof("Public Encryption Key:\n%s\n", org.Data.Body.PublicEncryptionKey)

	logger.Info("Tags for CAs:")
	for k, v := range index.Data.Body.Tags.CAForward {
		logger.Infof("  %s => %s\n", k, v)
	}
	logger.Info("CA tags:")
	for k, v := range index.Data.Body.Tags.CAReverse {
		logger.Infof("  %s => %s\n", k, v)
	}
	logger.Info("Tags for entities:")
	for k, v := range index.Data.Body.Tags.EntityForward {
		logger.Infof("  %s => %s\n", k, v)
	}
	logger.Info("Entity tags:")
	for k, v := range index.Data.Body.Tags.EntityReverse {
		logger.Infof("  %s => %s\n", k, v)
	}

	logger.Info("Pairing keys:")
	for k, v := range index.Data.Body.PairingKeys {
		logger.Infof("  %s => %s\n", k, v)
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

	logger.Info("Registering nodes")
	for {
		size, err := fsAPI.IncomingSize("registration")
		if err != nil {
			panic(logger.Errorf("Can't get queue size: %s", err))
		}
		logger.Infof("Found %d nodes to register\n", size)
		if size > 0 {
			regJson, err := fsAPI.PopIncoming("registration")
			if err != nil {
				panic(logger.Errorf("Can't pop registration: %s", err))
			}

			nodeReg, err := node.NewRegistration(regJson)
			if err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(logger.Errorf("Can't load registration: %s", err))
			}

			pairingId := nodeReg.Data.Options.PairingId
			pairingKey := indx.Data.Body.PairingKeys[pairingId]
			if err := nodeReg.Verify(pairingKey.Key); err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(logger.Errorf("Couldn't verify registration: %s", err))
			}

			node, err := node.NewFromRegistration(nodeReg)
			if err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(logger.Errorf("Couldn't create node from registration: %s", err))
			}
			//node.Data.Body.Tags = pairingKey.Tags
			indx.AddEntityTags(node.Data.Body.Id, pairingKey.Tags)
			// Create node container
			// Sign node container
			// Push signed container to node's incoming queue

			nodeContainer, err := document.NewContainer(nil)
			if err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(logger.Errorf("Couldn't create node container: %s", err))
			}

			nodeContainer.Data.Body = node.Dump()
			if err := org.Sign(nodeContainer); err != nil {
				fsAPI.PushIncoming(fsAPI.Id, "registration", regJson)
				panic(logger.Errorf("Couldn't sign node container: %s", err))

			}

			// For each tag, look for CAs
			for _, tag := range pairingKey.Tags {
				for _, caId := range indx.Data.Body.Tags.CAForward[tag] {

					// For each CA get a CSR for node
					csrContainerJson, err := fsAPI.PopOutgoing(node.Data.Body.Id, "csrs")
					if err != nil {
						panic(logger.Errorf("Couldn't get a csr: %s", err))
					}

					csrContainer, err := document.NewContainer(csrContainerJson)
					if err != nil {
						panic(logger.Errorf("Couldn't create container from json: %s", err))
					}

					if err := node.Verify(csrContainer); err != nil {
						panic(logger.Errorf("Couldn't verify CSR: %s", err))
					}

					csrJson := csrContainer.Data.Body
					csr, err := x509.NewCSR(csrJson)
					if err != nil {
						panic(logger.Errorf("Couldn't create csr from json: %s", err))
					}

					// Get the CA
					caContainerJson, err := fsAPI.GetPrivate(fsAPI.Id, caId)
					caContainer, err := document.NewContainer(caContainerJson)
					if err != nil {
						panic(logger.Errorf("Couldn't create container from json: %s", err))
					}
					caJson, err := org.VerifyThenDecrypt(caContainer)
					if err != nil {
						panic(logger.Errorf("Couldn't verify and decrypt ca container: %s", err))
					}

					ca, err := x509.NewCA(caJson)
					if err != nil {
						panic(logger.Errorf("Couldn't create ca: %s", err))
					}

					// Create a cert
					cert, err := ca.Sign(csr)
					if err != nil {
						panic(logger.Errorf("Couldn't sign csr: %s", err))
					}

					// Sign cert
					certContainer, err := document.NewContainer(nil)
					if err != nil {
						panic(logger.Errorf("Couldn't create cert container: %s", err))
					}
					certContainer.Data.Options.Source = org.Data.Body.Id
					certContainer.Data.Body = cert.Dump()
					if err := org.Sign(certContainer); err != nil {
						panic(logger.Errorf("Couldn't sign cert container: %s", err))
					}

					// Push cert to node's incoming queue
					if err := fsAPI.PushIncoming(node.Data.Body.Id, "certs", certContainer.Dump()); err != nil {
						panic(logger.Errorf("Couldn't push cert to node: %s", err))

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
Manages the Organisation.

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
