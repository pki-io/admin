package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/entity"
	"github.com/pki-io/pki.io/fs"
	"github.com/pki-io/pki.io/index"
	n "github.com/pki-io/pki.io/node"
	"github.com/pki-io/pki.io/x509"
)

const MinCSRs = 5

func LoadNodeIndex(fsAPI *fs.FsAPI, node *n.Node) *index.NodeIndex {
	indexJson, err := fsAPI.GetPrivate(node.Data.Body.Id, "index")
	if err != nil {
		panic(logger.Errorf("Could not load index data: %s", err))

	}
	indexContainer, err := document.NewContainer(indexJson)
	if err != nil {
		panic(logger.Errorf("Could not load index container: %s", err))
	}

	if err := node.Verify(indexContainer); err != nil {
		panic(logger.Errorf("Could not verify index: %s", err))
	}

	decryptedIndexJson, err := node.Decrypt(indexContainer)
	if err != nil {
		panic(logger.Errorf("Could not decrypt container: %s", err))
	}

	indx, err := index.NewNode(decryptedIndexJson)
	if err != nil {
		panic(logger.Errorf("Could not create indx: %s", err))
	}
	return indx
}

func SaveNodeIndex(fsAPI *fs.FsAPI, node *n.Node, indx *index.NodeIndex) {
	encryptedIndexContainer, err := node.EncryptThenSignString(indx.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt and sign index: %s", err))
	}
	if err := fsAPI.SendPrivate(node.Data.Body.Id, "index", encryptedIndexContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save encrypted: %s", err))
	}
}

func nodeNewCSR(fsAPI *fs.FsAPI, node *n.Node, org *entity.Entity) string {
	logger.Info("Creating new CSR")
	csr, err := x509.NewCSR(nil)
	if err != nil {
		panic(fmt.Sprintf("Could not generate CSR: %s", err))
	}

	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = node.Data.Body.Name
	csr.Generate()

	logger.Info("Saving local CSR")
	csrContainer, err := node.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt CSR: %s", err))
	}
	if err := fsAPI.SendPrivate(node.Data.Body.Id, csr.Data.Body.Id, csrContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save CSR: %s", err))
	}

	logger.Info("Pushing public CSR")
	csrPublic, err := csr.Public()
	if err != nil {
		panic(logger.Errorf("Could not get public CSR: %s", err))
	}

	csrPublicContainer, err := node.SignString(csrPublic.Dump())
	if err != nil {
		panic(logger.Errorf("Could not sign public CSR: %s", err))
	}

	if err := fsAPI.PushOutgoing("csrs", csrPublicContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not send public CSR: %s", err))
	}

	return csr.Data.Body.Id
}

func nodeGenerateCSRs(fsAPI *fs.FsAPI, node *n.Node, org *entity.Entity) []string {
	numCSRs, err := fsAPI.OutgoingSize(fsAPI.Id, "csrs")
	if err != nil {
		panic(logger.Errorf("Could not get csr queue size: %s", err))
	}

	csrs := make([]string, 0)
	for i := 0; i < MinCSRs-numCSRs; i++ {
		csrs = append(csrs, nodeNewCSR(fsAPI, node, org))
	}
	return csrs
}

func nodeNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)
	pairingId := argv["--pairing-id"].(string)
	pairingKey := argv["--pairing-key"].(string)
	offline := argv["--offline"].(bool)

	conf := LoadConfig()
	fsAPI := LoadAPI()
	admin := LoadAdmin(fsAPI, conf)
	org := LoadOrgPublic(fsAPI, admin, conf)

	logger.Info("Creating new node")
	node, err := n.New(nil)
	if err != nil {
		panic(logger.Errorf("Could not create node entity: %s:", err))
	}
	node.Data.Body.Name = name
	node.Data.Body.Id = NewID()

	// Acting on behalf of the node now, so all API is done for the node
	fsAPI.Id = node.Data.Body.Id

	logger.Info("Generating node keys")
	if err := node.GenerateKeys(); err != nil {
		panic(logger.Errorf("Could not generate node keys: %s", err))
	}

	logger.Info("Saving node config")
	conf.AddNode(node.Data.Body.Name, node.Data.Body.Id)
	if err := conf.Save(); err != nil {
		panic(logger.Errorf("Could not save admin config: %s", err))
	}

	var nodeJson string
	if offline {
		nodeJson = node.Dump()

	} else {
		nodeJson = node.DumpPublic()
		if err != nil {
			panic(logger.Errorf("Could not dump public node: %s:", err))
		}

	}
	entities := []*entity.Entity{org}
	container, err := node.EncryptThenAuthenticateString(nodeJson, entities, pairingId, pairingKey)
	if err != nil {
		panic(logger.Errorf("Could encrypt and authenticate node: %s:", err))
	}

	logger.Info("Encrypting and authenticating container")

	logger.Info("Pushing container to org")
	if err := fsAPI.PushIncoming(org.Data.Body.Id, "registration", container.Dump()); err != nil {
		panic(logger.Errorf("Could not push document to org: %s", err))
	}
	logger.Info("Saving node")
	if err := fsAPI.WriteLocal(node.Data.Body.Name, node.Dump()); err != nil {
		panic(logger.Errorf("Could not save node: %s", err))
	}

	logger.Info("Creating node index")
	indx, err := index.NewNode(nil)
	if err != nil {
		panic(logger.Errorf("Could not create index: %s", err))
	}

	// create crs
	logger.Info("Creating CSRs")
	nodeGenerateCSRs(fsAPI, node, org)

	logger.Info("Saving index")
	SaveNodeIndex(fsAPI, node, indx)
	return nil
}

func nodeProcessRegistrations(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI()
	admin := LoadAdmin(fsAPI, conf)
	org := LoadOrgPublic(fsAPI, admin, conf)

	// Need to rename the privte key files a bit. Then look them up via the config. Would do a search until
	// name is matchedc

	nodeJson, err := fsAPI.ReadLocal(name)
	if err != nil {
		panic(logger.Errorf("Could not read node file: %s", err))
	}

	node, err := n.New(nodeJson)
	if err != nil {
		panic(logger.Errorf("Could not load node json: %s", err))
	}

	fsAPI.Id = node.Data.Body.Id

	for {
		size, err := fsAPI.IncomingSize("registrations")
		if err != nil {
			panic(logger.Errorf("Can't get queue size: %s", err))
		}

		logger.Infof("Found %d registrations to process", size)
		if size > 0 {
			nodeContainerJson, err := fsAPI.PopIncoming("registrations")
			if err != nil {
				panic(logger.Errorf("Can't pop registrations: %s", err))
			}

			nodeContainer, err := document.NewContainer(nodeContainerJson)
			if err != nil {
				panic(logger.Errorf("Can't create node container: %s", err))
			}

			if err := org.Verify(nodeContainer); err != nil {
				panic(logger.Errorf("Can't verify node registration: %s", err))
			}

			nodeReg, err := n.New(nodeContainer.Data.Body)
			if err != nil {
				panic(logger.Errorf("Can't node registration: %s", err))
			}
			logger.Info(nodeReg)
		}
	}

	return nil
}

func nodeProcessCerts(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI()
	admin := LoadAdmin(fsAPI, conf)
	org := LoadOrgPublic(fsAPI, admin, conf)

	// Need to rename the privte key files a bit. Then look them up via the config. Would do a search until
	// name is matchedc

	nodeJson, err := fsAPI.ReadLocal(name)
	if err != nil {
		panic(logger.Errorf("Could not read node file: %s", err))
	}

	node, err := n.New(nodeJson)
	if err != nil {
		panic(logger.Errorf("Could not load node json: %s", err))
	}

	fsAPI.Id = node.Data.Body.Id
	indx := LoadNodeIndex(fsAPI, node)

	// For each income cert
	for {
		size, err := fsAPI.IncomingSize("certs")
		if err != nil {
			panic(logger.Errorf("Can't get queue size: %s", err))
		}

		logger.Infof("Found %d certs to process\n", size)
		if size > 0 {
			certContainerJson, err := fsAPI.PopIncoming("certs")
			if err != nil {
				panic(logger.Errorf("Can't pop cert: %s", err))
			}

			certContainer, err := document.NewContainer(certContainerJson)
			if err != nil {
				panic(logger.Errorf("Can't create cert container: %s", err))
			}
			if err := org.Verify(certContainer); err != nil {
				panic(logger.Errorf("Cert didn't verify: %s", err))
			}

			cert, err := x509.NewCertificate(certContainer.Data.Body)
			if err != nil {
				panic(logger.Errorf("Can't load certificate: %s", err))
			}

			// Read local CSR to get private key
			csrContainerJson, err := fsAPI.GetPrivate(fsAPI.Id, cert.Data.Body.Id)
			if err != nil {
				panic(logger.Errorf("Couldn't load csr file: %s", err))
			}

			csrContainer, err := document.NewContainer(csrContainerJson)
			if err != nil {
				panic(logger.Errorf("Couldn't load CSR container: %s", err))
			}

			csrJson, err := node.VerifyThenDecrypt(csrContainer)
			if err != nil {
				panic(logger.Errorf("Couldn't verify and decrypt container: %s", err))
			}

			csr, err := x509.NewCSR(csrJson)
			if err != nil {
				panic(logger.Errorf("Couldn't load csr: %s", err))
			}

			// Set the private key from the csr
			cert.Data.Body.PrivateKey = csr.Data.Body.PrivateKey

			// Reuse container
			updatedCertContainer, err := node.EncryptThenSignString(cert.Dump(), nil)
			if err != nil {
				panic(logger.Errorf("Could not encrypt then sign cert: %s", err))
			}

			if err := fsAPI.SendPrivate(node.Data.Body.Id, cert.Data.Body.Id, updatedCertContainer.Dump()); err != nil {
				panic(logger.Errorf("Could save cert: %s", err))
			}

			if err := indx.AddCertTags(cert.Data.Body.Id, cert.Data.Body.Tags); err != nil {
				panic(logger.Errorf("Could not add cert tags: %s", err))
			}

		} else {
			break
		}
	}
	logger.Info("Creating CSRs")
	nodeGenerateCSRs(fsAPI, node, org)

	logger.Info("Saving index")
	SaveNodeIndex(fsAPI, node, indx)
	return nil
}

func nodeShow(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)

	fsAPI := LoadAPI()

	nodeJson, err := fsAPI.ReadLocal(name)
	if err != nil {
		panic(logger.Errorf("Could not read node file: %s", err))
	}

	node, err := n.New(nodeJson)
	if err != nil {
		panic(logger.Errorf("Could not load node json: %s", err))
	}

	fsAPI.Id = node.Data.Body.Id
	indx := LoadNodeIndex(fsAPI, node)

	logger.Infof("Public Signing Key:\n%s", node.Data.Body.PublicSigningKey)
	logger.Infof("Public Encryption Key:\n%s", node.Data.Body.PublicEncryptionKey)
	logger.Infof("Certificate tags:\n%s", indx.Data.Body.Tags.CertForward)

	return nil
}

func nodeCert(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)
	inTags := argv["--tags"].(string)
	exportFile := argv["--export"] // Optional, so check for nil later

	fsAPI := LoadAPI()

	nodeJson, err := fsAPI.ReadLocal(name)
	if err != nil {
		panic(logger.Errorf("Could not read node file: %s", err))
	}

	node, err := n.New(nodeJson)
	if err != nil {
		panic(logger.Errorf("Could not load node json: %s", err))
	}

	fsAPI.Id = node.Data.Body.Id
	indx := LoadNodeIndex(fsAPI, node)

	var files []ExportFile
	for _, tag := range ParseTags(inTags) {
		logger.Infof("Getting certs for tag: %s", tag)
		for _, certId := range indx.Data.Body.Tags.CertForward[tag] {
			logger.Infof("Loading certificate: %s", certId)
			certContainerJson, err := fsAPI.GetPrivate(fsAPI.Id, certId)
			if err != nil {
				panic(logger.Errorf("Could not load cert container json: %s", err))
			}

			certContainer, err := document.NewContainer(certContainerJson)
			if err != nil {
				panic(logger.Errorf("Could not load cert container: %s", err))
			}

			certJson, err := node.VerifyThenDecrypt(certContainer)
			if err != nil {
				panic(logger.Errorf("Could not verify and decrypt: %s", err))
			}

			cert, err := x509.NewCertificate(certJson)
			if err != nil {
				panic(logger.Errorf("Could not load cert: %s", err))
			}
			switch exportFile.(type) {
			case nil:
				logger.Infof("Certificate:\n%s\n\n", cert.Data.Body.Certificate)
				logger.Infof("Private Key:\n%s\n\n", cert.Data.Body.PrivateKey)
			case string:
				files = append(files, ExportFile{Name: "cert.pem", Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
				files = append(files, ExportFile{Name: "key.pem", Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})
			}

		}
	}
	if len(files) > 0 {
		logger.Info("Exporting")
		Export(files, exportFile.(string))
	}
	return nil
}

// Node related commands
func runNode(args []string) (err error) {

	usage := `
Manages nodes.

Usage:
    pki.io node [--help]
    pki.io node new <name> --pairing-id=<id> --pairing-key=<key> [--offline]
    pki.io node run --name=<name>
    pki.io node show --name=<name>
    pki.io node cert --name=<name> --tags=<tags> [--export=<file>]

Options:
    --pairing-id=<id>   Pairing ID
    --pairing-key=<key> Pairing Key
    --name=<name>       Node name
    --offline           Create node in offline mode (false)
    --cert=<cert>       Certificate ID
    --export=<file>     Export data to file or "-" for STDOUT
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["new"].(bool) {
		nodeNew(argv)
	} else if argv["run"].(bool) {
		nodeProcessCerts(argv)
	} else if argv["show"].(bool) {
		nodeShow(argv)
	} else if argv["cert"].(bool) {
		nodeCert(argv)
	}
	return nil
}
