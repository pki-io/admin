// ThreatSpec package main
package main

import (
	"crypto/x509/pkix"
	"github.com/pki-io/core/config"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/entity"
	"github.com/pki-io/core/fs"
	"github.com/pki-io/core/index"
	"github.com/pki-io/core/node"
	"github.com/pki-io/core/x509"
	"os"
)

const (
	MinCSRs        int    = 5
	NodeConfigFile string = "node.conf"
)

type NodeApp struct {
	entities struct {
		node *node.Node
		org  *entity.Entity
	}
	config struct {
		node *config.NodeConfig
	}
	fs struct {
		local *fs.Local
		home  *fs.Home
		api   *fs.Api
	}
	index struct {
		node *index.NodeIndex
	}
}

func NewNodeApp() *NodeApp {
	return new(NodeApp)
}

func (app *NodeApp) InitLocalFs() {
	var err error
	app.fs.local, err = fs.NewLocal(os.Getenv("PKIIO_LOCAL"))
	checkAppFatal("Couldn't initialize local fs: %s", err)
}

func (app *NodeApp) InitHomeFs() {
	var err error
	app.fs.home, err = fs.NewHome(os.Getenv("PKIIO_HOME"))
	checkAppFatal("Couldn't initialise home fs: %s", err)
}

func (app *NodeApp) InitApiFs() {
	var err error
	app.fs.api, err = fs.NewAPI(app.fs.local.Path)
	checkAppFatal("Couldn't initialise api fs: %s", err)
}

func (app *NodeApp) LoadNodeConfig() {
	var err error
	nodeConfig, err := app.fs.local.Read(NodeConfigFile)
	checkAppFatal("Couldn't read node config: %s", err)
	app.config.node, err = config.NewNode()
	checkAppFatal("Couldn't initialize org config: %s", err)
	err = app.config.node.Load(nodeConfig)
	checkAppFatal("Couldn't load node config: %s", err)
}

func (app *NodeApp) Load() {
	// Copy/paste *all* the things!

	logger.Info("Loading node app")

	app.InitLocalFs()
	app.LoadNodeConfig()
	app.InitHomeFs()
	app.InitApiFs()
}

func (app *NodeApp) CreateNodeIndex() {
	var err error
	logger.Info("Creating node index")
	app.index.node, err = index.NewNode(nil)
	checkAppFatal("Could not create index: %s", err)
	app.index.node.Data.Body.Id = NewID()
}

func (app *NodeApp) LoadNodeIndex() {
	logger.Info("Loading node index")
	var err error

	if app.entities.node == nil {
		checkAppFatal("Node not found in app: %s", err)
	}

	nodeConfig, err := app.config.node.GetNode(app.entities.node.Data.Body.Name)
	checkAppFatal("Could not get id for node: %s", err)

	indexJson, err := app.fs.api.GetPrivate(app.entities.node.Data.Body.Id, nodeConfig.Index)
	checkAppFatal("Could not get index container: %s", err)

	indexContainer, err := document.NewContainer(indexJson)
	checkAppFatal("Could not load index container: %s", err)

	err = app.entities.node.Verify(indexContainer)
	checkAppFatal("Could not verify index: %s", err)

	decryptedIndexJson, err := app.entities.node.Decrypt(indexContainer)
	checkAppFatal("Could not decrypt container: %s", err)

	app.index.node, err = index.NewNode(decryptedIndexJson)
	checkAppFatal("Could not create index: %s", err)
}

func (app *NodeApp) SaveNodeIndex() {
	logger.Info("Saving node index")
	var err error
	encryptedIndexContainer, err := app.entities.node.EncryptThenSignString(app.index.node.Dump(), nil)
	checkAppFatal("Could not encrypt and sign index: %s", err)
	err = app.fs.api.SendPrivate(app.entities.node.Data.Body.Id, app.index.node.Data.Body.Id, encryptedIndexContainer.Dump())
	checkAppFatal("Could not save encrypted: %s", err)
}

func (app *NodeApp) CreateNodeConfig() {
	logger.Info("Creating node config")
	var err error
	app.config.node, err = config.NewNode()
	checkAppFatal("Couldn't initialize node config: %s", err)
	err = app.config.node.AddNode(app.entities.node.Data.Body.Name, app.entities.node.Data.Body.Id, app.index.node.Data.Body.Id)
	checkUserFatal("Cannot add node to node config: %s", err)
}

func (app *NodeApp) SaveNodeConfig() {
	cfgString, err := app.config.node.Dump()
	checkAppFatal("Couldn't dump node config: %s", err)

	err = app.fs.local.Write(NodeConfigFile, cfgString)
	checkAppFatal("Couldn't save node config: %s", err)
}

func (app *NodeApp) GenerateCSRs() {
	logger.Info("Generating CSRs")

	numCSRs, err := app.fs.api.OutgoingSize(app.fs.api.Id, "csrs")
	checkAppFatal("Could not get csr queue size: %s", err)

	for i := 0; i < MinCSRs-numCSRs; i++ {
		app.NewCSR()
	}
}

// ThreatSpec TMv0.1 for NodeApp.NewCSR
// Creates a new node CSR for App:Node

func (app *NodeApp) NewCSR() {
	logger.Info("Creating new CSR")
	csr, err := x509.NewCSR(nil)
	checkAppFatal("Could not generate CSR: %s", err)

	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = app.entities.node.Data.Body.Name
	subject := pkix.Name{CommonName: csr.Data.Body.Name}
	csr.Generate(&subject)

	logger.Info("Saving local CSR")

	csrContainer, err := app.entities.node.EncryptThenSignString(csr.Dump(), nil)
	checkAppFatal("Could not encrypt CSR: %s", err)

	err = app.fs.api.SendPrivate(app.entities.node.Data.Body.Id, csr.Data.Body.Id, csrContainer.Dump())
	checkAppFatal("Could not save CSR: %s", err)

	logger.Info("Pushing public CSR")
	csrPublic, err := csr.Public()
	checkAppFatal("Could not get public CSR: %s", err)

	csrPublicContainer, err := app.entities.node.SignString(csrPublic.Dump())
	checkAppFatal("Could not sign public CSR: %s", err)

	err = app.fs.api.PushOutgoing("csrs", csrPublicContainer.Dump())
	checkAppFatal("Could not send public CSR: %s", err)
}

func (app *NodeApp) ProcessNextCert() {
	certContainerJson, err := app.fs.api.PopIncoming("certs")
	checkAppFatal("Can't pop cert: %s", err)

	certContainer, err := document.NewContainer(certContainerJson)
	checkAppFatal("Can't create cert container: %s", err)

	err = app.entities.org.Verify(certContainer)
	checkAppFatal("Cert didn't verify: %s", err)

	cert, err := x509.NewCertificate(certContainer.Data.Body)
	checkAppFatal("Can't load certificate: %s", err)

	csrContainerJson, err := app.fs.api.GetPrivate(app.entities.node.Data.Body.Id, cert.Data.Body.Id)
	checkAppFatal("Couldn't load csr file: %s", err)

	csrContainer, err := document.NewContainer(csrContainerJson)
	checkAppFatal("Couldn't load CSR container: %s", err)

	csrJson, err := app.entities.node.VerifyThenDecrypt(csrContainer)
	checkAppFatal("Couldn't verify and decrypt container: %s", err)

	csr, err := x509.NewCSR(csrJson)
	checkAppFatal("Couldn't load csr: %s", err)

	// Set the private key from the csr
	cert.Data.Body.PrivateKey = csr.Data.Body.PrivateKey

	updatedCertContainer, err := app.entities.node.EncryptThenSignString(cert.Dump(), nil)
	checkAppFatal("Could not encrypt then sign cert: %s", err)

	err = app.fs.api.SendPrivate(app.entities.node.Data.Body.Id, cert.Data.Body.Id, updatedCertContainer.Dump())
	checkAppFatal("Could save cert: %s", err)

	err = app.index.node.AddCertTags(cert.Data.Body.Id, cert.Data.Body.Tags)
	checkAppFatal("Could not add cert tags: %s", err)
}

func (app *NodeApp) ProcessCerts() {
	logger.Info("Processing certs")
	for {
		size, err := app.fs.api.IncomingSize("certs")
		checkAppFatal("Can't get queue size: %s", err)
		logger.Infof("Found %d certs to process", size)

		if size > 0 {
			app.ProcessNextCert()
		} else {
			break
		}
	}
}

func (app *NodeApp) CreateNodeEntity(name string) {
	var err error
	logger.Info("Creating new node")
	app.entities.node, err = node.New(nil)
	checkAppFatal("Could not create node entity: %s:", err)
	app.entities.node.Data.Body.Name = name
	app.entities.node.Data.Body.Id = NewID()

	logger.Info("Generating node keys")
	err = app.entities.node.GenerateKeys()
	checkAppFatal("Could not generate node keys: %s", err)
}

func (app *NodeApp) SecureSendStringToOrg(nodeJson, pairingId, pairingKey string) {
	logger.Info("Encrypting node for org")

	container, err := app.entities.org.EncryptThenAuthenticateString(nodeJson, pairingId, pairingKey)
	checkAppFatal("Could encrypt and authenticate node: %s:", err)

	logger.Info("Pushing container to org")
	err = app.fs.api.PushIncoming(app.entities.org.Data.Body.Id, "registration", container.Dump())
	checkAppFatal("Could not push document to org: %s", err)
}

func (app *NodeApp) SecureSendPublicToOrg(pairingId, pairingKey string) {
	nodeJson := app.entities.node.DumpPublic()
	app.SecureSendStringToOrg(nodeJson, pairingId, pairingKey)
}

func (app *NodeApp) SecureSendPrivateToOrg(pairingId, pairingKey string) {
	nodeJson := app.entities.node.Dump()
	app.SecureSendStringToOrg(nodeJson, pairingId, pairingKey)
}

func (app *NodeApp) GetCertificates(tags string) []*x509.Certificate {
	var certs []*x509.Certificate
	for _, tag := range ParseTags(tags) {
		logger.Infof("Getting certs for tag: %s", tag)
		for _, certId := range app.index.node.Data.Body.Tags.CertForward[tag] {
			cert := app.GetCertificate(certId)
			certs = append(certs, cert)
		}
	}
	return certs
}

func (app *NodeApp) GetCertificate(id string) *x509.Certificate {
	certContainerJson, err := app.fs.api.GetPrivate(app.entities.node.Data.Body.Id, id)
	checkAppFatal("Could not load cert container json: %s", err)

	certContainer, err := document.NewContainer(certContainerJson)
	checkAppFatal("Could not load cert container: %s", err)

	certJson, err := app.entities.node.VerifyThenDecrypt(certContainer)
	checkAppFatal("Could not verify and decrypt: %s", err)

	cert, err := x509.NewCertificate(certJson)
	checkAppFatal("Could not load cert: %s", err)
	return cert
}
