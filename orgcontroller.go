package main

import (
	"fmt"
	"github.com/pki-io/core/config"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/entity"
	"github.com/pki-io/core/index"
	"github.com/pki-io/core/node"
	"github.com/pki-io/core/x509"
)

const (
	OrgConfigFile string = "org.conf"
)

type OrgParams struct {
	org           *string
	admin         *string
	confirmDelete *string
	private       *bool
}

type OrgController struct {
	env    *Environment
	config *config.OrgConfig
	org    *entity.Entity
}

func NewOrgParams() *OrgParams {
	return new(OrgParams)
}

func (params *OrgParams) ValidateOrg() error {
	if *params.org == "" {
		return fmt.Errorf("Invalid org: Cannot be empty")
	}
	return nil
}

func (params *OrgParams) ValidateAdmin() error {
	if *params.admin == "" {
		return fmt.Errorf("Invalid admin: Cannot be empty")
	}
	return nil
}

func NewOrgController(env *Environment) (*OrgController, error) {
	cont := new(OrgController)
	cont.env = env

	return cont, nil
}

func (cont *OrgController) LoadConfig() error {
	cont.env.logger.Debug("Loading org config")

	var err error
	if cont.config == nil {
		cont.config, err = config.NewOrg()
		if err != nil {
			return err
		}
	}

	exists, err := cont.env.fs.local.Exists(OrgConfigFile)
	if err != nil {
		return err
	}

	if exists {
		orgConfig, err := cont.env.fs.local.Read(OrgConfigFile)
		if err != nil {
			return err
		}

		err = cont.config.Load(orgConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cont *OrgController) SaveConfig() error {
	cont.env.logger.Debug("Saving org config")

	cfgString, err := cont.config.Dump()
	if err != nil {
		return err
	}

	if err := cont.env.fs.local.Write(OrgConfigFile, cfgString); err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) CreateOrg(name string) error {
	cont.env.logger.Debug("Creating org")

	var err error
	cont.org, err = entity.New(nil)
	if err != nil {
		return err
	}

	cont.org.Data.Body.Id = NewID()
	cont.org.Data.Body.Name = name

	if err := cont.org.GenerateKeys(); err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) GetOrgAdmins() ([]*entity.Entity, error) {
	cont.env.logger.Debug("Getting org admins")

	index, err := cont.GetIndex()
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Getting admins from index")
	adminIds, err := index.GetAdmins()
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Loading admin entities")
	adminEntities := make([]*entity.Entity, 0, 0)
	for _, id := range adminIds {
		admin, err := cont.env.controllers.admin.GetAdmin(id)
		if err != nil {
			return nil, err
		}
		adminEntities = append(adminEntities, admin)
	}

	return adminEntities, nil
}

func (cont *OrgController) LoadPublicOrg() error {
	cont.env.logger.Debug("Loading public org")

	orgId := cont.config.Data.Id

	cont.env.logger.Debugf("Reading org %s", orgId)
	orgPublicJson, err := cont.env.fs.home.Read(orgId)
	if err != nil {
		return err
	}

	cont.org, err = entity.New(orgPublicJson)
	if err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) SavePublicOrg() error {
	cont.env.logger.Debug("Saving public org")

	if err := cont.env.fs.home.Write(cont.org.Data.Body.Id, cont.org.DumpPublic()); err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) SavePrivateOrg() error {
	cont.env.logger.Debug("Saving private org")

	admins, err := cont.GetOrgAdmins()
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Encrypting org for admins")
	container, err := cont.org.EncryptThenSignString(cont.org.Dump(), admins)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Sending encrypted org")
	if err := cont.env.api.SendPrivate(cont.org.Data.Body.Id, cont.org.Data.Body.Id, container.Dump()); err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) LoadPrivateOrg() error {
	cont.env.logger.Debug("Loading private org")

	orgId := cont.config.Data.Id

	cont.env.logger.Debugf("Loading private org '%s'", orgId)
	orgEntity, err := cont.env.api.GetPrivate(orgId, orgId)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Creating new org container")
	contontainer, err := document.NewContainer(orgEntity)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Verifying container")
	err = cont.org.Verify(contontainer)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Decrypting container")
	decryptedOrgJson, err := cont.env.controllers.admin.admin.Decrypt(contontainer)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Creating org entity")
	cont.org, err = entity.New(decryptedOrgJson)
	if err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) CreateIndex() (*index.OrgIndex, error) {
	cont.env.logger.Debug("Creating new org index")

	index, err := index.NewOrg(nil)
	if err != nil {
		return nil, err
	}

	index.Data.Body.Id = NewID()
	index.Data.Body.ParentId = cont.org.Data.Body.Id

	return index, nil
}

func (cont *OrgController) GetIndex() (*index.OrgIndex, error) {
	cont.env.logger.Debug("Getting org index")

	orgIndexId := cont.config.Data.Index
	indexJson, err := cont.env.api.GetPrivate(cont.org.Data.Body.Id, orgIndexId)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Creating container for index")
	indexContainer, err := document.NewContainer(indexJson)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Verifying container")
	err = cont.org.Verify(indexContainer)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Decrypting container")
	decryptedIndexJson, err := cont.org.Decrypt(indexContainer)
	if err != nil {
		return nil, err
	}

	cont.env.logger.Debug("Creating new index")
	index, err := index.NewOrg(decryptedIndexJson)
	if err != nil {
		return nil, err
	}

	return index, nil
}

func (cont *OrgController) SaveIndex(index *index.OrgIndex) error {
	cont.env.logger.Debug("Saving org index")

	encryptedIndexContainer, err := cont.org.EncryptThenSignString(index.Dump(), nil)
	if err != nil {
		return err
	}

	err = cont.env.api.SendPrivate(cont.org.Data.Body.Id, index.Data.Body.Id, encryptedIndexContainer.Dump())
	if err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) GetCA(id string) (*x509.CA, error) {
	cont.env.logger.Debug("Getting CA")

	org := cont.env.controllers.org.org
	caContainerJson, err := cont.env.api.GetPrivate(org.Data.Body.Id, id)
	if err != nil {
		return nil, err
	}

	caContainer, err := document.NewContainer(caContainerJson)
	if err != nil {
		return nil, err
	}

	caJson, err := org.VerifyThenDecrypt(caContainer)
	if err != nil {
		return nil, err
	}

	ca, err := x509.NewCA(caJson)
	if err != nil {
		return nil, err
	}

	return ca, nil
}

func (cont *OrgController) SignCSR(node *node.Node, caId, tag string) error {
	cont.env.logger.Debug("Signing CSR for node")

	csrContainerJson, err := cont.env.api.PopOutgoing(node.Data.Body.Id, "csrs")
	if err != nil {
		return err
	}

	csrContainer, err := document.NewContainer(csrContainerJson)
	if err != nil {
		return err
	}

	if err := node.Verify(csrContainer); err != nil {
		return err
	}

	csrJson := csrContainer.Data.Body
	csr, err := x509.NewCSR(csrJson)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Setting CSR name from node")
	csr.Data.Body.Name = node.Data.Body.Name

	ca, err := cont.GetCA(caId)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Creating certificate")
	cert, err := ca.Sign(csr)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Tagging certificate")
	cert.Data.Body.Tags = append(cert.Data.Body.Tags, tag)

	cont.env.logger.Debug("Signing certificate")
	certContainer, err := document.NewContainer(nil)
	if err != nil {
		return err
	}

	org := cont.env.controllers.org.org
	certContainer.Data.Options.Source = org.Data.Body.Id
	certContainer.Data.Body = cert.Dump()
	if err := org.Sign(certContainer); err != nil {
		return err
	}

	cont.env.logger.Debug("Pushing certificate to node")
	if err := cont.env.api.PushIncoming(node.Data.Body.Id, "certs", certContainer.Dump()); err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) RegisterNextNode() error {
	cont.env.logger.Debug("Registering next node")

	org := cont.env.controllers.org.org

	regJson, err := cont.env.api.PopIncoming(org.Data.Body.Id, "registration")
	if err != nil {
		return err
	}

	container, err := document.NewContainer(regJson)
	if err != nil {
		cont.env.logger.Warn("Unable to create container from registration json. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Data.Body.Id, "registration", regJson)
		return err
	}

	pairingId := container.Data.Options.SignatureInputs["key-id"]
	cont.env.logger.Debugf("Reading pairing key for '%s'", pairingId)

	cont.env.logger.Debug("Getting org index")
	index, err := cont.GetIndex()
	if err != nil {
		return err
	}

	pairingKey, ok := index.Data.Body.PairingKeys[pairingId]
	if !ok {
		cont.env.logger.Warn("Unable to find pairing key. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Data.Body.Id, "registration", regJson)
		return fmt.Errorf("Could not find pairing key for '%s'", pairingId)
	}

	cont.env.logger.Debug("Verifying and decrypting node registration")
	nodeJson, err := org.VerifyAuthenticationThenDecrypt(container, pairingKey.Key)
	if err != nil {
		cont.env.logger.Warn("Unable to decrypt node registration. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Data.Body.Id, "registration", regJson)
		return err
	}

	node, err := node.New(nodeJson)
	if err != nil {
		cont.env.logger.Warn("Unable to create node. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Data.Body.Id, "registration", regJson)
		return err
	}

	cont.env.logger.Debug("Adding node to index")
	index.AddEntityTags(node.Data.Body.Id, pairingKey.Tags)
	index.AddNode(node.Data.Body.Name, node.Data.Body.Id)

	cont.env.logger.Debug("Encrypting and signing node for org")
	nodeContainer, err := org.EncryptThenSignString(node.Dump(), nil)
	if err != nil {
		cont.env.logger.Warn("Unable to encrypt node for org. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Data.Body.Id, "registration", regJson)
		return err
	}

	cont.env.logger.Debug("Sending node to org")
	if err := cont.env.api.SendPrivate(org.Data.Body.Id, node.Data.Body.Id, nodeContainer.Dump()); err != nil {
		cont.env.logger.Warn("Unable to send node to org. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Data.Body.Id, "registration", regJson)
		return err
	}

	for _, tag := range pairingKey.Tags {
		cont.env.logger.Debugf("Looking for CAs for tag '%s'", tag)
		for _, caId := range index.Data.Body.Tags.CAForward[tag] {
			cont.env.logger.Debugf("Found CA '%s'", caId)
			if err := cont.SignCSR(node, caId, tag); err != nil {
				return err
			}
		}
	}

	cont.env.logger.Debug("Saving index")
	if err := cont.SaveIndex(index); err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) RegisterNodes() error {
	cont.env.logger.Debug("Registering nodes")

	org := cont.env.controllers.org.org

	for {
		size, err := cont.env.api.IncomingSize(org.Data.Body.Id, "registration")
		if err != nil {
			return err
		}

		cont.env.logger.Debugf("Found %d nodes to register", size)

		if size > 0 {
			if err := cont.RegisterNextNode(); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (cont *OrgController) Init(params *OrgParams) error {

	cont.env.logger.Debug("Initializing new org")

	cont.env.logger.Debug("Validating parameters")
	if err := params.ValidateOrg(); err != nil {
		return err
	}

	if err := params.ValidateAdmin(); err != nil {
		return err
	}

	cont.env.logger.Debug("Initializing file system")

	if err := cont.env.LoadLocalFs(); err != nil {
		return err
	}

	if err := cont.env.LoadHomeFs(); err != nil {
		return err
	}

	cont.env.logger.Debug("Checking whether org directory exists")

	exists, err := cont.env.fs.local.Exists(*params.org)
	if err != nil {
		return err
	}

	if exists {
		cont.env.logger.Warnf("Org directory '%s' already exists", *params.org)
		return nil
	}

	cont.env.logger.Debug("Initializing the admin controller")

	cont.env.controllers.admin, err = NewAdminController(cont.env)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Loading admin config")

	if err := cont.env.controllers.admin.LoadConfig(); err != nil {
		return err
	}

	cont.env.logger.Debug("Looking for duplicate orgs")

	if cont.env.controllers.admin.config.OrgExists(*params.org) {
		cont.env.logger.Warnf("Org already exists: %s", *params.org)
		return nil
	}

	cont.env.logger.Debug("Creating org directory")

	if err := cont.env.fs.local.CreateDirectory(*params.org); err != nil {
		return err
	}

	// Make all further fs calls relative to the Org
	cont.env.logger.Debug("Changing to org directory")

	if err := cont.env.fs.local.ChangeToDirectory(*params.org); err != nil {
		return err
	}

	// Initialize the API
	cont.env.logger.Debug("Initializing API")

	if err := cont.env.LoadAPI(); err != nil {
		return nil
	}

	cont.env.logger.Debug("Creating admin entity")

	if err := cont.env.controllers.admin.CreateAdmin(*params.admin); err != nil {
		return err
	}

	admin := cont.env.controllers.admin.admin

	cont.env.logger.Debug("Creating org entity")

	if err := cont.CreateOrg(*params.org); err != nil {
		return err
	}

	cont.env.logger.Debug("Loading org config")

	if err := cont.LoadConfig(); err != nil {
		return err
	}

	cont.env.logger.Debug("Saving admin")

	if err := cont.env.controllers.admin.SaveAdmin(); err != nil {
		return err
	}

	cont.env.logger.Debug("Creating org index")

	orgIndex, err := cont.CreateIndex()
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Adding admin to org index")

	orgIndex.AddAdmin(admin.Data.Body.Name, admin.Data.Body.Id)

	cont.env.logger.Debug("Saving org index")

	if err := cont.SaveIndex(orgIndex); err != nil {
		return err
	}

	cont.env.logger.Debug("Saving org config")

	cont.config.Data.Index = orgIndex.Data.Body.Id
	cont.config.Data.Id = cont.org.Data.Body.Id
	cont.config.Data.Name = cont.org.Data.Body.Name

	if err := cont.SaveConfig(); err != nil {
		return err
	}

	cont.env.logger.Debug("Saving public org")

	if err := cont.SavePublicOrg(); err != nil {
		return err
	}

	cont.env.logger.Debug("Adding org to public config")

	err = cont.env.controllers.admin.config.AddOrg(cont.org.Data.Body.Name, cont.org.Data.Body.Id, admin.Data.Body.Id)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Saving admin config")

	if err := cont.env.controllers.admin.SaveConfig(); err != nil {
		return err
	}

	cont.env.logger.Debug("Saving org")

	if err := cont.SavePrivateOrg(); err != nil {
		return err
	}

	return nil
}

func (cont *OrgController) List(params *OrgParams) ([]*entity.Entity, error) {

	var err error
	cont.env.logger.Debug("Loading home filesystem")
	if err := cont.env.LoadHomeFs(); err != nil {
		return nil, err
	}

	if cont.env.controllers.admin == nil {
		cont.env.controllers.admin, err = NewAdminController(cont.env)
		if err != nil {
			return nil, err
		}
	}

	cont.env.logger.Debug("Loading admin config")

	if err := cont.env.controllers.admin.LoadConfig(); err != nil {
		return nil, err
	}

	orgs := make([]*entity.Entity, 0)
	for _, org := range cont.env.controllers.admin.config.GetOrgs() {
		fmt.Printf("* %s %s\n", org.Name, org.Id)
		o, err := entity.New(nil)
		if err != nil {
			return nil, err
		}

		o.Data.Body.Id = org.Id
		o.Data.Body.Name = org.Name
		orgs = append(orgs, o)
	}

	return orgs, nil
}

func (cont *OrgController) Show(params *OrgParams) (*entity.Entity, error) {
	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	return cont.env.controllers.org.org, nil

}

func (cont *OrgController) RunEnv(params *OrgParams) error {
	if err := cont.RegisterNodes(); err != nil {
		return err
	}
	return nil
}

func (cont *OrgController) Run(params *OrgParams) error {
	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	// LoadAdminEnv has actually loaded a fresh new org controller
	// inside our current env, so all further actions need to be relative
	// to that controller. So we run RunEnv scoped to the correct environment.
	return cont.env.controllers.org.RunEnv(params)
}

func (cont *OrgController) Delete(params *OrgParams) error {

	return fmt.Errorf("Not implemented")
}
