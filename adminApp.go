// ThreatSpec package main
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

// ThreatSpec TMv0.1 for AdminApp.InitLocalFs
// Does initialisation of local filesytem for App:Admin

func (app *AdminApp) InitLocalFs() {
	var err error
	app.fs.local, err = fs.NewLocal(os.Getenv("PKIIO_LOCAL"))
	checkAppFatal("Couldn't initialize local fs: %s", err)
}

func (app *AdminApp) InitApiFs() {
	var err error
	app.fs.api, err = fs.NewAPI(app.fs.local.Path)
	checkAppFatal("Couldn't initialise api fs: %s", err)
}

func (app *AdminApp) InitHomeFs() {
	var err error
	app.fs.home, err = fs.NewHome(os.Getenv("PKIIO_HOME"))
	checkAppFatal("Couldn't initialise home fs: %s", err)
}

// ThreatSpec TMv0.1 for AdminApp.CreateOrgDirectory
// Creates org directory for App:Admin

func (app *AdminApp) CreateOrgDirectory(name string) {
	exists, err := app.fs.local.Exists(name)
	checkAppFatal("Couldn't check existence of org: %s", err)

	if exists {
		checkUserFatal("Org directory '%name' already exists.", name)
	}

	err = app.fs.local.CreateDirectory(name)
	checkAppFatal("Couldn't create org directory: %s", err)

	err = app.fs.local.ChangeToDirectory(name)
	checkAppFatal("Couldn't change to org directory: %s", err)
}

// ThreatSpec TMv0.1 for AdminApp.CreateAdminEntity
// Does org entity creation for App:Admin

func (app *AdminApp) CreateAdminEntity(name string) {
	var err error
	logger.Info("Creating Admin entity")
	app.entities.admin, err = entity.New(nil)
	checkAppFatal("Could not create admin entity: %s", err)

	app.entities.admin.Data.Body.Id = NewID()
	app.entities.admin.Data.Body.Name = name

	logger.Info("Generating admin keys")
	err = app.entities.admin.GenerateKeys()
	checkAppFatal("Could not generate admin keys: %s", err)
}

// ThreatSpec TMv0.1 for AdminApp.CreateOrgEntity
// Does org entity creation for App:Org

func (app *AdminApp) CreateOrgEntity(name string) {
	var err error
	logger.Info("Creating Org entity")
	app.entities.org, err = entity.New(nil)
	checkAppFatal("Could not create org entity: %s", err)

	app.entities.org.Data.Body.Id = NewID()
	app.entities.org.Data.Body.Name = name

	logger.Info("Generating Org keys")
	err = app.entities.org.GenerateKeys()
	checkAppFatal("Could not generate org keys: %s", err)
}

func (app *AdminApp) SaveAdminEntity() {
	id := app.entities.admin.Data.Body.Id

	logger.Info("Saving local admin")
	err := app.fs.home.Write(id, app.entities.admin.Dump())
	checkAppFatal("Couldn't save local admin: %s", err)

	logger.Info("Saving public admin")
	err = app.fs.api.SendPublic(id, id, app.entities.admin.DumpPublic())
	checkAppFatal("Couldn't save public admin: %s", err)
}

func (app *AdminApp) SaveOrgEntityPublic() {
	logger.Info("Saving org public entity to home")
	err := app.fs.home.Write(app.entities.org.Data.Body.Id, app.entities.org.DumpPublic())
	checkAppFatal("Couldn't save public org: %s", err)
}

func (app *AdminApp) LoadAdminEntity() {
	var err error
	logger.Info("Loading admin entity")

	orgName := app.config.org.Data.Name

	adminOrgConfig, err := app.config.admin.GetOrg(orgName)
	checkAppFatal("Could not get org from admin config: %s", err)

	adminId := adminOrgConfig.AdminId

	adminEntity, err := app.fs.home.Read(adminId)
	checkAppFatal("Couldn't read admin entity: %s", err)

	app.entities.admin, err = entity.New(adminEntity)
	checkAppFatal("Couldn't load admin entity: %s", err)
}

func (app *AdminApp) CreateAdminConfig() {
	var err error
	app.config.admin, err = config.NewAdmin()
	checkAppFatal("Couldn't initialize admin config: %s", err)

}

func (app *AdminApp) SaveAdminConfig() {
	logger.Info("Saving admin config")
	cfgString, err := app.config.admin.Dump()
	checkAppFatal("Couldn't dump admin config: %s", err)

	err = app.fs.home.Write(AdminConfigFile, cfgString)
	checkAppFatal("Couldn't save admin config: %s", err)
}

func (app *AdminApp) AdminConfigExists() (bool, error) {
	return app.fs.home.Exists(AdminConfigFile)
}

func (app *AdminApp) LoadAdminConfig() {
	logger.Info("Loading admin config")
	adminConfig, err := app.fs.home.Read(AdminConfigFile)
	checkAppFatal("Couldn't read admin config: %s", err)

	app.config.admin, err = config.NewAdmin()
	checkAppFatal("Couldn't initialize admin config: %s", err)

	err = app.config.admin.Load(adminConfig)
	checkAppFatal("Couldn't load admin config: %s", err)
}

// ThreatSpec TMv0.1 for AdminApp.SendOrgEntity
// It encrypts and uploads org entity for App:Org

func (app *AdminApp) SendOrgEntity() {

	// Get array of admin entities
	//entities := []*entity.Entity{app.entities.admin}
	container, err := app.entities.org.EncryptThenSignString(app.entities.org.Dump(), app.GetAdminEntities())
	checkAppFatal("Couldn't encrypt org: %s", err)

	err = app.fs.api.Authenticate(app.entities.org.Data.Body.Id, "")
	checkAppFatal("Couldn't authenticate to FS api: %s", err)

	err = app.fs.api.StorePrivate(app.entities.org.Data.Body.Id, container.Dump())
	checkAppFatal("Couldn't store container to json: %s", err)
}

func (app *AdminApp) LoadOrgEntity() {
	logger.Info("Loading org entity")
	var err error

	orgId := app.config.org.Data.Id
	app.fs.api.Authenticate(orgId, "")

	orgPublicJson, err := app.fs.home.Read(orgId)
	checkAppFatal("Couldn't read org public entity: %s", err)

	app.entities.org, err = entity.New(orgPublicJson)
	checkAppFatal("Could not create org entity: %s", err)

	orgEntity, err := app.fs.api.LoadPrivate(orgId)
	checkAppFatal("Couldn't load org entity: %s", err)

	orgContainer, err := document.NewContainer(orgEntity)
	checkAppFatal("Could not load org container: %s", err)

	err = app.entities.org.Verify(orgContainer)
	checkAppFatal("Could not verify org: %s", err)

	decryptedOrgJson, err := app.entities.admin.Decrypt(orgContainer)
	checkAppFatal("Could not decrypt container: %s", err)

	app.entities.org, err = entity.New(decryptedOrgJson)
	checkAppFatal("Could not create org entity: %s", err)
}

func (app *AdminApp) CreateOrgConfig() {
	var err error
	app.config.org, err = config.NewOrg()
	checkAppFatal("Couldn't initialize org config: %s", err)

	app.config.org.Data.Id = app.entities.org.Data.Body.Id
	app.config.org.Data.Name = app.entities.org.Data.Body.Name
}

func (app *AdminApp) SaveOrgConfig() {
	cfgString, err := app.config.org.Dump()
	checkAppFatal("Couldn't dump admin config: %s", err)

	err = app.fs.local.Write(OrgConfigFile, cfgString)
	checkAppFatal("Could not store container to json: %s", err)
}

func (app *AdminApp) LoadOrgConfig() {
	orgConfig, err := app.fs.local.Read(OrgConfigFile)
	checkAppFatal("Couldn't read org config: %s", err)

	app.config.org, err = config.NewOrg()
	checkAppFatal("Couldn't initialize org config: %s", err)

	err = app.config.org.Load(orgConfig)
	checkAppFatal("Couldn't load org config: %s", err)
}

func (app *AdminApp) CreateOrgIndex() {
	var err error
	logger.Info("Creating org index")
	app.index.org, err = index.NewOrg(nil)
	checkAppFatal("Could not create index: %s", err)

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
	checkAppFatal("Could not load index container: %s", err)

	err = app.entities.org.Verify(indexContainer)
	checkAppFatal("Could not verify index: %s", err)

	decryptedIndexJson, err := app.entities.org.Decrypt(indexContainer)
	checkAppFatal("Could not decrypt container: %s", err)

	app.index.org, err = index.NewOrg(decryptedIndexJson)
	checkAppFatal("Could not create index: %s", err)
}

func (app *AdminApp) SaveOrgIndex() {
	var err error
	encryptedIndexContainer, err := app.entities.org.EncryptThenSignString(app.index.org.Dump(), nil)
	checkAppFatal("Could not encrypt and sign index: %s", err)

	err = app.fs.api.SendPrivate(app.entities.org.Data.Body.Id, app.index.org.Data.Body.Id, encryptedIndexContainer.Dump())
	checkAppFatal("Could not save encrypted: %s", err)
}

func (app *AdminApp) GetCA(id string) *x509.CA {
	logger.Info("Getting CA")
	caContainerJson, err := app.fs.api.GetPrivate(app.entities.org.Data.Body.Id, id)
	caContainer, err := document.NewContainer(caContainerJson)
	checkAppFatal("Couldn't create container from json: %s", err)

	caJson, err := app.entities.org.VerifyThenDecrypt(caContainer)
	checkAppFatal("Couldn't verify and decrypt ca container: %s", err)

	ca, err := x509.NewCA(caJson)
	checkAppFatal("Couldn't create ca: %s", err)
	return ca
}

func (app *AdminApp) SignCSRForNode(node *node.Node, caId, tag string) {
	logger.Info("Getting CSR for node")
	csrContainerJson, err := app.fs.api.PopOutgoing(node.Data.Body.Id, "csrs")
	checkAppFatal("Couldn't get a csr: %s", err)

	csrContainer, err := document.NewContainer(csrContainerJson)
	checkAppFatal("Couldn't create container from json: %s", err)

	err = node.Verify(csrContainer)
	checkAppFatal("Couldn't verify CSR: %s", err)

	csrJson := csrContainer.Data.Body
	csr, err := x509.NewCSR(csrJson)
	checkAppFatal("Couldn't create csr from json: %s", err)

	logger.Info("Setting CSR name from node")
	csr.Data.Body.Name = node.Data.Body.Name

	ca := app.GetCA(caId)

	logger.Info("Creating certificate")
	cert, err := ca.Sign(csr)
	checkAppFatal("Couldn't sign csr: %s", err)

	logger.Info("Tagging cert")
	cert.Data.Body.Tags = append(cert.Data.Body.Tags, tag)

	logger.Info("Signing cert")
	certContainer, err := document.NewContainer(nil)
	checkAppFatal("Couldn't create cert container: %s", err)

	certContainer.Data.Options.Source = app.entities.org.Data.Body.Id
	certContainer.Data.Body = cert.Dump()
	err = app.entities.org.Sign(certContainer)
	checkAppFatal("Couldn't sign cert container: %s", err)

	logger.Info("Pushing certificate to node")
	err = app.fs.api.PushIncoming(node.Data.Body.Id, "certs", certContainer.Dump())
	checkAppFatal("Couldn't push cert to node: %s", err)
}

func (app *AdminApp) RegisterNextNode() {
	orgId := app.entities.org.Data.Body.Id

	regJson, err := app.fs.api.PopIncoming("registration")
	checkAppFatal("Can't pop registration: %s", err)

	container, err := document.NewContainer(regJson)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "registration", regJson)
		checkAppFatal("Can't load registration: %s", err)
	}
	pairingId := container.Data.Options.SignatureInputs["key-id"]
	logger.Infof("Reading pairing key: %s", pairingId)
	pairingKey := app.index.org.Data.Body.PairingKeys[pairingId]

	logger.Info("Verifying and decrypting node registration")
	nodeJson, err := app.entities.org.VerifyAuthenticationThenDecrypt(container, pairingKey.Key)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "registration", regJson)
		checkAppFatal("Couldn't verify then decrypt registration: %s", err)
	}

	node, err := node.New(nodeJson)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "registration", regJson)
		checkAppFatal("Couldn't create node from registration: %s", err)
	}

	logger.Info("Adding node to index")
	app.index.org.AddEntityTags(node.Data.Body.Id, pairingKey.Tags)
	app.index.org.AddNode(node.Data.Body.Name, node.Data.Body.Id)

	logger.Info("Encrypting and signing node for Org")
	nodeContainer, err := app.entities.org.EncryptThenSignString(node.Dump(), nil)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "registration", regJson)
		checkAppFatal("Couldn't encrypt then sign node: %s", err)
	}

	if err := app.fs.api.SendPrivate(orgId, node.Data.Body.Id, nodeContainer.Dump()); err != nil {
		checkAppFatal("Could not save node: %s", err)
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
		checkAppFatal("Can't get queue size: %s", err)

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
	checkAppFatal("Couldn't get node config: %s", err)

	nodeContainerJson, err := app.fs.api.GetPrivate(app.entities.org.Data.Body.Id, nodeId)
	checkAppFatal("Couldn't get node container: %s", err)

	nodeContainer, err := document.NewContainer(nodeContainerJson)
	checkAppFatal("Could not load node container: %s", err)

	nodeJson, err := app.entities.org.VerifyThenDecrypt(nodeContainer)
	checkAppFatal("Couldn't get verify then decrypt node: %s", err)

	nde, err := node.New(nodeJson)
	checkAppFatal("Couldn't create node: %s", err)
	return nde
}

func (app *AdminApp) SecureSendStringToOrg(adminJson, inviteId, inviteKey string) {
	logger.Info("Encrypting node for org")

	container, err := app.entities.admin.EncryptThenAuthenticateString(adminJson, inviteId, inviteKey)
	checkAppFatal("Could encrypt and authenticate invite: %s:", err)

	logger.Info("Pushing container to org")
	err = app.fs.api.PushIncoming(app.config.org.Data.Id, "invite", container.Dump())
	checkAppFatal("Could not push document to org: %s", err)
}

func (app *AdminApp) SecureSendPublicToOrg(inviteId, inviteKey string) {
	adminJson := app.entities.admin.DumpPublic()
	app.SecureSendStringToOrg(adminJson, inviteId, inviteKey)
}

func (app *AdminApp) SecureSendPrivateToOrg(inviteId, inviteKey string) {
	adminJson := app.entities.admin.Dump()
	app.SecureSendStringToOrg(adminJson, inviteId, inviteKey)
}

func (app *AdminApp) GetAdminEntity(id string) *entity.Entity {
	adminJson, err := app.fs.api.GetPublic(id, id)
	checkAppFatal("Couldn't get admin json: %s", err)

	admin, err := entity.New(adminJson)
	checkAppFatal("Couldn't create admin entity: %s", err)

	return admin
}

func (app *AdminApp) GetAdminEntities() []*entity.Entity {
	adminIds, err := app.index.org.GetAdmins()
	checkAppFatal("Can't get admins from index: %s", err)

	adminEntities := make([]*entity.Entity, 0, 0)
	for _, id := range adminIds {
		adminEntities = append(adminEntities, app.GetAdminEntity(id))
	}

	return adminEntities
}

func (app *AdminApp) ProcessNextInvite() {
	orgId := app.entities.org.Data.Body.Id

	inviteJson, err := app.fs.api.PopIncoming("invite")
	checkAppFatal("Can't pop invite: %s", err)

	container, err := document.NewContainer(inviteJson)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "invite", inviteJson)
		checkAppFatal("Can't load invite: %s", err)
	}

	inviteId := container.Data.Options.SignatureInputs["key-id"]
	logger.Infof("Reading invite key: %s", inviteId)
	inviteKey, err := app.index.org.GetInviteKey(inviteId)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "invite", inviteJson)
		checkAppFatal("Couldn't get invite key: %s", err)
	}

	logger.Info("Verifying and decrypting admin invite")
	adminJson, err := app.entities.org.VerifyAuthenticationThenDecrypt(container, inviteKey.Key)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "invite", inviteJson)
		checkAppFatal("Couldn't verify then decrypt invite: %s", err)
	}

	admin, err := entity.New(adminJson)
	if err != nil {
		app.fs.api.PushIncoming(orgId, "invite", inviteJson)
		checkAppFatal("Couldn't load admin entity: %s", err)
	}

	err = app.index.org.AddAdmin(admin.Data.Body.Name, admin.Data.Body.Id)
	checkAppFatal("Couldn't add admin to index: %s", err)

	app.SendOrgEntity()
	//app.entities.org.EncryptThenSignString(app.entities.org.Dump(), app.GetAdminEntities())

	orgContainer, err := app.entities.admin.EncryptThenAuthenticateString(app.entities.org.DumpPublic(), inviteId, inviteKey.Key)
	checkAppFatal("Couldn't encrypt and authenticate org public: %s", err)

	err = app.fs.api.PushIncoming(admin.Data.Body.Id, "invite", orgContainer.Dump())
	checkAppFatal("Could not push org container: %s", err)

	// Delete invite ID
}

func (app *AdminApp) ProcessInvites() {
	logger.Info("Processing invites")
	for {
		size, err := app.fs.api.IncomingSize("invite")
		checkAppFatal("Can't get queue size: %s", err)

		logger.Infof("Found %d invites to process", size)

		if size > 0 {
			app.ProcessNextInvite()
		} else {
			break
		}
	}
}

func (app *AdminApp) CompleteInvite(inviteId, inviteKey string) {
	err := app.fs.api.Authenticate(app.entities.admin.Data.Body.Id, "")
	checkAppFatal("Couldn't authenticate to API: %s", err)

	orgContainerJson, err := app.fs.api.PopIncoming("invite")
	checkAppFatal("Can't pop invite: %s", err)

	orgContainer, err := document.NewContainer(orgContainerJson)
	checkAppFatal("Can't create org container: %s", err)

	orgJson, err := app.entities.admin.VerifyAuthenticationThenDecrypt(orgContainer, inviteKey)
	checkAppFatal("Couldn't verify invite: %s", err)

	app.entities.org, err = entity.New(orgJson)
	checkAppFatal("Couldn't create org entitiy: %s", err)

	app.SaveOrgEntityPublic()
}
