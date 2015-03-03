package main

import (
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
	if err != nil {
		panic(logger.Errorf("Couldn't initialize local fs: %s", err))
	}
}

func (app *NodeApp) InitHomeFs() {
	var err error
	app.fs.home, err = fs.NewHome(os.Getenv("PKIIO_HOME"))
	if err != nil {
		panic(logger.Errorf("Couldn't initialise home fs: %s", err))
	}
}

func (app *NodeApp) InitApiFs() {
	var err error
	app.fs.api, err = fs.NewAPI(app.fs.local.Path)
	if err != nil {
		panic(logger.Errorf("Couldn't initialise api fs: %s", err))
	}
}

func (app *NodeApp) LoadNodeConfig() {
	var err error
	nodeConfig, err := app.fs.local.Read(NodeConfigFile)
	if err != nil {
		panic(logger.Errorf("Couldn't read node config: %s", err))
	}
	app.config.node, err = config.NewNode()
	if err != nil {
		panic(logger.Errorf("Couldn't initialize org config: %s", err))
	}
	if err := app.config.node.Load(nodeConfig); err != nil {
		panic(logger.Errorf("Couldn't load node config: %s", err))
	}
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
	if err != nil {
		panic(logger.Errorf("Could not create index: %s", err))
	}
	app.index.node.Data.Body.Id = NewID()
}

func (app *NodeApp) LoadNodeIndex() {
	logger.Info("Loading node index")
	var err error

	if app.entities.node == nil {
		panic(logger.Errorf("Node not found in app: %s", err))

	}

	nodeConfig, err := app.config.node.GetNode(app.entities.node.Data.Body.Name)
	if err != nil {
		panic(logger.Errorf("Could not get id for node: %s", err))

	}
	indexJson, err := app.fs.api.GetPrivate(app.entities.node.Data.Body.Id, nodeConfig.Index)
	if err != nil {
		panic(logger.Errorf("Could not get index container: %s", err))
	}

	indexContainer, err := document.NewContainer(indexJson)
	if err != nil {
		panic(logger.Errorf("Could not load index container: %s", err))
	}

	if err := app.entities.node.Verify(indexContainer); err != nil {
		panic(logger.Errorf("Could not verify index: %s", err))
	}

	decryptedIndexJson, err := app.entities.node.Decrypt(indexContainer)
	if err != nil {
		panic(logger.Errorf("Could not decrypt container: %s", err))
	}

	app.index.node, err = index.NewNode(decryptedIndexJson)
	if err != nil {
		panic(logger.Errorf("Could not create index: %s", err))
	}
}

func (app *NodeApp) SaveNodeIndex() {
	logger.Info("Saving node index")
	var err error
	encryptedIndexContainer, err := app.entities.node.EncryptThenSignString(app.index.node.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt and sign index: %s", err))
	}
	if err := app.fs.api.SendPrivate(app.entities.node.Data.Body.Id, app.index.node.Data.Body.Id, encryptedIndexContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save encrypted: %s", err))
	}
}

func (app *NodeApp) CreateNodeConfig() {
	logger.Info("Creating node config")
	var err error
	app.config.node, err = config.NewNode()
	if err != nil {
		panic(logger.Errorf("Couldn't initialize node config: %s", err))
	}
	app.config.node.AddNode(app.entities.node.Data.Body.Name, app.entities.node.Data.Body.Id, app.index.node.Data.Body.Id)
}

func (app *NodeApp) SaveNodeConfig() {
	cfgString, err := app.config.node.Dump()
	if err != nil {
		panic(logger.Errorf("Couldn't dump node config: %s", err))
	}

	if err := app.fs.local.Write(NodeConfigFile, cfgString); err != nil {
		panic(logger.Errorf("Couldn't save node config: %s", err))
	}
}

func (app *NodeApp) GenerateCSRs() {
	logger.Info("Generating CSRs")

	numCSRs, err := app.fs.api.OutgoingSize(app.fs.api.Id, "csrs")
	if err != nil {
		panic(logger.Errorf("Could not get csr queue size: %s", err))
	}

	for i := 0; i < MinCSRs-numCSRs; i++ {
		app.NewCSR()
	}
}

func (app *NodeApp) NewCSR() {
	logger.Info("Creating new CSR")
	csr, err := x509.NewCSR(nil)
	if err != nil {
		panic(logger.Errorf("Could not generate CSR: %s", err))
	}

	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = app.entities.node.Data.Body.Name
	csr.Generate()

	logger.Info("Saving local CSR")

	csrContainer, err := app.entities.node.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt CSR: %s", err))
	}

	if err := app.fs.api.SendPrivate(app.entities.node.Data.Body.Id, csr.Data.Body.Id, csrContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save CSR: %s", err))
	}

	logger.Info("Pushing public CSR")
	csrPublic, err := csr.Public()
	if err != nil {
		panic(logger.Errorf("Could not get public CSR: %s", err))
	}

	csrPublicContainer, err := app.entities.node.SignString(csrPublic.Dump())
	if err != nil {
		panic(logger.Errorf("Could not sign public CSR: %s", err))
	}

	if err := app.fs.api.PushOutgoing("csrs", csrPublicContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not send public CSR: %s", err))
	}
}

func (app *NodeApp) ProcessNextCert() {
	certContainerJson, err := app.fs.api.PopIncoming("certs")
	if err != nil {
		panic(logger.Errorf("Can't pop cert: %s", err))
	}

	certContainer, err := document.NewContainer(certContainerJson)
	if err != nil {
		panic(logger.Errorf("Can't create cert container: %s", err))
	}

	if err := app.entities.org.Verify(certContainer); err != nil {
		panic(logger.Errorf("Cert didn't verify: %s", err))
	}

	cert, err := x509.NewCertificate(certContainer.Data.Body)
	if err != nil {
		panic(logger.Errorf("Can't load certificate: %s", err))
	}

	csrContainerJson, err := app.fs.api.GetPrivate(app.entities.node.Data.Body.Id, cert.Data.Body.Id)
	if err != nil {
		panic(logger.Errorf("Couldn't load csr file: %s", err))
	}

	csrContainer, err := document.NewContainer(csrContainerJson)
	if err != nil {
		panic(logger.Errorf("Couldn't load CSR container: %s", err))
	}

	csrJson, err := app.entities.node.VerifyThenDecrypt(csrContainer)
	if err != nil {
		panic(logger.Errorf("Couldn't verify and decrypt container: %s", err))
	}

	csr, err := x509.NewCSR(csrJson)
	if err != nil {
		panic(logger.Errorf("Couldn't load csr: %s", err))
	}

	// Set the private key from the csr
	cert.Data.Body.PrivateKey = csr.Data.Body.PrivateKey

	updatedCertContainer, err := app.entities.node.EncryptThenSignString(cert.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt then sign cert: %s", err))
	}

	if err := app.fs.api.SendPrivate(app.entities.node.Data.Body.Id, cert.Data.Body.Id, updatedCertContainer.Dump()); err != nil {
		panic(logger.Errorf("Could save cert: %s", err))
	}

	if err := app.index.node.AddCertTags(cert.Data.Body.Id, cert.Data.Body.Tags); err != nil {
		panic(logger.Errorf("Could not add cert tags: %s", err))
	}
}

func (app *NodeApp) ProcessCerts() {
	logger.Info("Processing certs")
	for {
		size, err := app.fs.api.IncomingSize("certs")
		if err != nil {
			panic(logger.Errorf("Can't get queue size: %s", err))
		}
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
	if err != nil {
		panic(logger.Errorf("Could not create node entity: %s:", err))
	}
	app.entities.node.Data.Body.Name = name
	app.entities.node.Data.Body.Id = NewID()

	logger.Info("Generating node keys")
	if err := app.entities.node.GenerateKeys(); err != nil {
		panic(logger.Errorf("Could not generate node keys: %s", err))
	}
}

func (app *NodeApp) SecureSendStringToOrg(nodeJson, pairingId, pairingKey string) {
	logger.Info("Encrypting node for org")
	entities := []*entity.Entity{app.entities.org}

	container, err := app.entities.org.EncryptThenAuthenticateString(nodeJson, entities, pairingId, pairingKey)
	if err != nil {
		panic(logger.Errorf("Could encrypt and authenticate node: %s:", err))
	}

	logger.Info("Pushing container to org")
	if err := app.fs.api.PushIncoming(app.entities.org.Data.Body.Id, "registration", container.Dump()); err != nil {
		panic(logger.Errorf("Could not push document to org: %s", err))
	}
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
	if err != nil {
		panic(logger.Errorf("Could not load cert container json: %s", err))
	}

	certContainer, err := document.NewContainer(certContainerJson)
	if err != nil {
		panic(logger.Errorf("Could not load cert container: %s", err))
	}

	certJson, err := app.entities.node.VerifyThenDecrypt(certContainer)
	if err != nil {
		panic(logger.Errorf("Could not verify and decrypt: %s", err))
	}

	cert, err := x509.NewCertificate(certJson)
	if err != nil {
		panic(logger.Errorf("Could not load cert: %s", err))
	}
	return cert
}
