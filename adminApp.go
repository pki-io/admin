package main

import (
	"github.com/pki-io/pki.io/config"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/entity"
	"github.com/pki-io/pki.io/fs"
	"github.com/pki-io/pki.io/index"
	"github.com/pki-io/pki.io/node"
	"github.com/pki-io/pki.io/x509"
	"os"
)

const (
	AdminConfigFile string = "admin.conf"
	OrgConfigFile   string = "org.conf"
)

type AdminApp struct {
	entities struct {
		admin *entity.Entity
		org   *entity.Entity
	}
	config struct {
		admin *config.AdminConfig
		org   *config.OrgConfig
	}
	fs struct {
		local *fs.Local
		home  *fs.Home
		api   *fs.Api
	}
	index struct {
		org *index.OrgIndex
	}
}

func NewAdminApp() *AdminApp {
	return new(AdminApp)
}

func (app *AdminApp) InitLocalFs() {
	var err error
	app.fs.local, err = fs.NewLocal(os.Getenv("PKIIO_LOCAL"))
	if err != nil {
		panic(logger.Errorf("Couldn't initialize local fs: %s", err))
	}
}

func (app *AdminApp) InitApiFs() {
	var err error
	app.fs.api, err = fs.NewAPI(app.fs.local.Path)
	if err != nil {
		panic(logger.Errorf("Couldn't initialise api fs: %s", err))
	}
}

func (app *AdminApp) InitHomeFs() {
	var err error
	app.fs.home, err = fs.NewHome(os.Getenv("PKIIO_HOME"))
	if err != nil {
		panic(logger.Errorf("Couldn't initialise home fs: %s", err))
	}
}

func (app *AdminApp) CreateOrgDirectory(name string) {
	exists, err := app.fs.local.Exists(name)
	if err != nil {
		panic(logger.Errorf("Couldn't check existence of org: %s", err))
	}

	if exists {
		panic(logger.Error("Org directory already exists"))
	}

	if err := app.fs.local.CreateDirectory(name); err != nil {
		panic(logger.Errorf("Couldn't create org directory: %s", err))
	}
	if err := app.fs.local.ChangeToDirectory(name); err != nil {
		panic(logger.Errorf("Couldn't change to org directory: %s", err))
	}
}

func (app *AdminApp) CreateAdminEntity(name string) {
	var err error
	logger.Info("Creating Admin entity")
	app.entities.admin, err = entity.New(nil)
	if err != nil {
		panic(logger.Errorf("Could not create admin entity: %s", err))
	}

	app.entities.admin.Data.Body.Id = NewID()
	app.entities.admin.Data.Body.Name = name

	logger.Info("Generating admin keys")
	err = app.entities.admin.GenerateKeys()
	if err != nil {
		panic(logger.Errorf("Could not generate admin keys: %s", err))
	}
}

func (app *AdminApp) CreateOrgEntity(name string) {
	var err error
	logger.Info("Creating Org entity")
	app.entities.org, err = entity.New(nil)
	if err != nil {
		panic(logger.Errorf("Could not create org entity: %s", err))
	}
	app.entities.org.Data.Body.Id = NewID()
	app.entities.org.Data.Body.Name = name

	logger.Info("Generating Org keys")
	err = app.entities.org.GenerateKeys()
	if err != nil {
		panic(logger.Errorf("Could not generate org keys: %s", err))
	}
}

func (app *AdminApp) SaveAdminEntity() {
	logger.Info("Saving admin")
	if err := app.fs.home.Write(app.entities.admin.Data.Body.Id, app.entities.admin.Dump()); err != nil {
		panic(logger.Errorf("Couldn't save admin: %s", err))
	}
}

func (app *AdminApp) LoadAdminEntity() {
	var err error
	logger.Info("Loading admin entity")

	orgName := app.config.org.Data.Name

	adminOrgConfig, err := app.config.admin.GetOrg(orgName)

	adminId := adminOrgConfig.AdminId

	adminEntity, err := app.fs.home.Read(adminId)
	if err != nil {
		panic(logger.Errorf("Couldn't read admin entity: %s", err))
	}

	app.entities.admin, err = entity.New(adminEntity)
	if err != nil {
		panic(logger.Errorf("Couldn't load admin entity: %s", err))
	}
}

func (app *AdminApp) CreateAdminConfig() {
	var err error
	app.config.admin, err = config.NewAdmin()
	if err != nil {
		panic(logger.Errorf("Couldn't initialize admin config: %s", err))
	}
	app.config.admin.AddOrg(app.entities.org.Data.Body.Name, app.entities.org.Data.Body.Id, app.entities.admin.Data.Body.Id)
}

func (app *AdminApp) SaveAdminConfig() {
	cfgString, err := app.config.admin.Dump()
	if err != nil {
		panic(logger.Errorf("Couldn't dump admin config: %s", err))
	}

	if err := app.fs.home.Write(AdminConfigFile, cfgString); err != nil {
		panic(logger.Errorf("Couldn't save admin config: %s", err))
	}
}

func (app *AdminApp) LoadAdminConfig() {
	adminConfig, err := app.fs.home.Read(AdminConfigFile)
	if err != nil {
		panic(logger.Errorf("Couldn't read admin config: %s", err))
	}
	app.config.admin, err = config.NewAdmin()
	if err != nil {
		panic(logger.Errorf("Couldn't initialize admin config: %s", err))
	}
	if err := app.config.admin.Load(adminConfig); err != nil {
		panic(logger.Errorf("Couldn't load admin config: %s", err))
	}
}

func (app *AdminApp) SaveOrgEntity() {
	container, err := app.entities.admin.EncryptThenSignString(app.entities.org.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Couldn't encrypt org: %s", err))
	}

	if err := app.fs.api.Authenticate(app.entities.org.Data.Body.Id, ""); err != nil {
		panic(logger.Errorf("Couldn't authenticate to FS api: %s", err))
	}

	if err := app.fs.api.StorePrivate(app.entities.org.Data.Body.Id, container.Dump()); err != nil {
		panic(logger.Errorf("Couldn't store container to json: %s", err))
	}
}

func (app *AdminApp) LoadOrgEntity() {
	logger.Info("Loading org entity")
	var err error

	orgId := app.config.org.Data.Id
	app.fs.api.Authenticate(orgId, "")

	orgEntity, err := app.fs.api.LoadPrivate(orgId)
	if err != nil {
		panic(logger.Errorf("Couldn't load org entity: %s", err))
	}

	orgContainer, err := document.NewContainer(orgEntity)
	if err != nil {
		panic(logger.Errorf("Could not load org container: %s", err))
	}

	if err := app.entities.admin.Verify(orgContainer); err != nil {
		panic(logger.Errorf("Could not verify org: %s", err))
	}

	decryptedOrgJson, err := app.entities.admin.Decrypt(orgContainer)
	if err != nil {
		panic(logger.Errorf("Could not decrypt container: %s", err))
	}

	app.entities.org, err = entity.New(decryptedOrgJson)
	if err != nil {
		panic(logger.Errorf("Could not create org entity: %s", err))
	}
}

func (app *AdminApp) CreateOrgConfig() {
	var err error
	app.config.org, err = config.NewOrg()
	if err != nil {
		panic(logger.Errorf("Couldn't initialize org config: %s", err))
	}
	app.config.org.Data.Id = app.entities.org.Data.Body.Id
	app.config.org.Data.Name = app.entities.org.Data.Body.Name
}

func (app *AdminApp) SaveOrgConfig() {
	cfgString, err := app.config.org.Dump()
	if err != nil {
		panic(logger.Errorf("Couldn't dump admin config: %s", err))
	}
	if err := app.fs.local.Write(OrgConfigFile, cfgString); err != nil {
		panic(logger.Errorf("Could not store container to json: %s", err))
	}
}

func (app *AdminApp) LoadOrgConfig() {
	orgConfig, err := app.fs.local.Read(OrgConfigFile)
	if err != nil {
		panic(logger.Errorf("Couldn't read org config: %s", err))
	}
	app.config.org, err = config.NewOrg()
	if err != nil {
		panic(logger.Errorf("Couldn't initialize org config: %s", err))
	}
	if err := app.config.org.Load(orgConfig); err != nil {
		panic(logger.Errorf("Couldn't load org config: %s", err))
	}
}

func (app *AdminApp) CreateOrgIndex() {
	var err error
	logger.Info("Creating org index")
	app.index.org, err = index.NewOrg(nil)
	if err != nil {
		panic(logger.Errorf("Could not create index: %s", err))
	}
	app.index.org.Data.Body.Id = NewID()
	app.index.org.Data.Body.ParentId = app.entities.org.Data.Body.Id
}

func (app *AdminApp) Load() {
	logger.Info("Loading admin app")

	app.InitLocalFs()
	app.LoadOrgConfig()
	app.InitHomeFs()
	app.LoadAdminConfig()
	app.InitApiFs()
	app.LoadAdminEntity()
	app.LoadOrgEntity()
}

func (app *AdminApp) LoadOrgIndex() {
	var err error

	orgIndexId := app.config.org.Data.Index
	indexJson, err := app.fs.api.GetPrivate(app.entities.org.Data.Body.Id, orgIndexId)

	indexContainer, err := document.NewContainer(indexJson)
	if err != nil {
		panic(logger.Errorf("Could not load index container: %s", err))
	}

	if err := app.entities.org.Verify(indexContainer); err != nil {
		panic(logger.Errorf("Could not verify index: %s", err))
	}

	decryptedIndexJson, err := app.entities.org.Decrypt(indexContainer)
	if err != nil {
		panic(logger.Errorf("Could not decrypt container: %s", err))
	}

	app.index.org, err = index.NewOrg(decryptedIndexJson)
	if err != nil {
		panic(logger.Errorf("Could not create index: %s", err))
	}
}

func (app *AdminApp) SaveOrgIndex() {
	var err error
	encryptedIndexContainer, err := app.entities.org.EncryptThenSignString(app.index.org.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt and sign index: %s", err))
	}
	if err := app.fs.api.SendPrivate(app.entities.org.Data.Body.Id, app.index.org.Data.Body.Id, encryptedIndexContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save encrypted: %s", err))
	}
}

func (app *AdminApp) SignCSRForNode(node *node.Node, caId, tag string) {
	logger.Info("Getting CSR for node")
	csrContainerJson, err := app.fs.api.PopOutgoing(node.Data.Body.Id, "csrs")
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

	logger.Info("Getting CA")
	caContainerJson, err := app.fs.api.GetPrivate(app.entities.org.Data.Body.Id, caId)
	caContainer, err := document.NewContainer(caContainerJson)
	if err != nil {
		panic(logger.Errorf("Couldn't create container from json: %s", err))
	}
	caJson, err := app.entities.org.VerifyThenDecrypt(caContainer)
	if err != nil {
		panic(logger.Errorf("Couldn't verify and decrypt ca container: %s", err))
	}

	ca, err := x509.NewCA(caJson)
	if err != nil {
		panic(logger.Errorf("Couldn't create ca: %s", err))
	}

	logger.Info("Creating certificate")
	cert, err := ca.Sign(csr)
	if err != nil {
		panic(logger.Errorf("Couldn't sign csr: %s", err))
	}

	logger.Info("Tagging cert")
	cert.Data.Body.Tags = append(cert.Data.Body.Tags, tag)

	logger.Info("Signing cert")
	certContainer, err := document.NewContainer(nil)
	if err != nil {
		panic(logger.Errorf("Couldn't create cert container: %s", err))
	}
	certContainer.Data.Options.Source = app.entities.org.Data.Body.Id
	certContainer.Data.Body = cert.Dump()
	if err := app.entities.org.Sign(certContainer); err != nil {
		panic(logger.Errorf("Couldn't sign cert container: %s", err))
	}

	logger.Info("Pushing certificate to node")
	if err := app.fs.api.PushIncoming(node.Data.Body.Id, "certs", certContainer.Dump()); err != nil {
		panic(logger.Errorf("Couldn't push cert to node: %s", err))

	}
}

func (app *AdminApp) RegisterNextNode() {
	orgId := app.entities.org.Data.Body.Id

	regJson, err := app.fs.api.PopIncoming("registration")
	if err != nil {
		panic(logger.Errorf("Can't pop registration: %s", err))
	}

	container, err := document.NewContainer(regJson)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "registration", regJson)
		panic(logger.Errorf("Can't load registration: %s", err))
	}
	pairingId := container.Data.Options.SignatureInputs["key-id"]
	logger.Infof("Reading pairing key: %s", pairingId)
	pairingKey := app.index.org.Data.Body.PairingKeys[pairingId]

	logger.Info("Verifying and decrypting node registration")
	nodeJson, err := app.entities.org.VerifyAuthenticationThenDecrypt(container, pairingKey.Key)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "registration", regJson)
		panic(logger.Errorf("Couldn't verify then decrypt registration: %s", err))
	}

	node, err := node.New(nodeJson)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "registration", regJson)
		panic(logger.Errorf("Couldn't create node from registration: %s", err))
	}

	logger.Info("Adding node to index")
	app.index.org.AddEntityTags(node.Data.Body.Id, pairingKey.Tags)
	app.index.org.AddNode(node.Data.Body.Name, node.Data.Body.Id)

	logger.Info("Encrypting and signing node for Org")
	nodeContainer, err := app.entities.org.EncryptThenSignString(node.Dump(), nil)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "registration", regJson)
		panic(logger.Errorf("Couldn't encrypt then sign node: %s", err))
	}

	if err := app.fs.api.SendPrivate(orgId, node.Data.Body.Id, nodeContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save node: %s", err))
	}

	for _, tag := range pairingKey.Tags {
		logger.Infof("Looking for CAs for tag %s", tag)
		for _, caId := range app.index.org.Data.Body.Tags.CAForward[tag] {
			logger.Infof("Found CA %s", caId)
			app.SignCSRForNode(node, caId, tag)
		}
	}

}

func (app *AdminApp) RegisterNodes() {
	logger.Info("Registering nodes")
	for {
		size, err := app.fs.api.IncomingSize("registration")
		if err != nil {
			panic(logger.Errorf("Can't get queue size: %s", err))
		}
		logger.Infof("Found %d nodes to register", size)

		if size > 0 {
			app.RegisterNextNode()
		} else {
			break
		}
	}
}

func (app *AdminApp) GetNode(name string) *node.Node {
	nodeId, err := app.index.org.GetNode(name)
	if err != nil {
		panic(logger.Errorf("Couldn't get node config: %s", err))
	}

	nodeContainerJson, err := app.fs.api.GetPrivate(app.entities.org.Data.Body.Id, nodeId)
	if err != nil {
		panic(logger.Errorf("Couldn't get node container: %s", err))
	}

	nodeContainer, err := document.NewContainer(nodeContainerJson)
	if err != nil {
		panic(logger.Errorf("Could not load node container: %s", err))
	}

	nodeJson, err := app.entities.org.VerifyThenDecrypt(nodeContainer)
	if err != nil {
		panic(logger.Errorf("Couldn't get verify then decrypt node: %s", err))
	}

	nde, err := node.New(nodeJson)
	if err != nil {
		panic(logger.Errorf("Couldn't create node: %s", err))
	}
	return nde
}
