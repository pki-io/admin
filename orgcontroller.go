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
		return fmt.Errorf("invalid org: Cannot be empty")
	}
	return nil
}

func (params *OrgParams) ValidateAdmin() error {
	if *params.admin == "" {
		return fmt.Errorf("invalid admin: Cannot be empty")
	}
	return nil
}

func NewOrgController(env *Environment) (*OrgController, error) {
	cont := new(OrgController)
	cont.env = env

	return cont, nil
}

func (cont *OrgController) OrgId() string {
	logger.Trace("returning org id")
	return cont.org.Id()
}

func (cont *OrgController) LoadConfig() error {
	logger.Debug("loading org config")

	var err error
	if cont.config == nil {
		cont.config, err = config.NewOrg()
		if err != nil {
			return err
		}
	}

	logger.Debugf("checking if local org config '%s' exists", OrgConfigFile)
	exists, err := cont.env.fs.local.Exists(OrgConfigFile)
	if err != nil {
		return err
	}

	if exists {
		logger.Debug("reading local org config")
		orgConfig, err := cont.env.fs.local.Read(OrgConfigFile)
		if err != nil {
			return err
		}

		err = cont.config.Load(orgConfig)
		if err != nil {
			return err
		}
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) SaveConfig() error {
	logger.Debug("saving org config")

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
	logger.Debug("creating org")

	var err error
	cont.org, err = entity.New(nil)
	if err != nil {
		return err
	}

	cont.org.Data.Body.Id = NewID()
	cont.org.Data.Body.Name = name

	logger.Debug("generating keys")
	if err := cont.org.GenerateKeys(); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) GetOrgAdmins() ([]*entity.Entity, error) {
	logger.Debug("getting org admins")

	index, err := cont.GetIndex()
	if err != nil {
		return nil, err
	}

	adminIds, err := index.GetAdmins()
	if err != nil {
		return nil, err
	}

	adminEntities := make([]*entity.Entity, 0, 0)
	for _, id := range adminIds {
		admin, err := cont.env.controllers.admin.GetAdmin(id)
		if err != nil {
			return nil, err
		}
		adminEntities = append(adminEntities, admin)
	}

	logger.Trace("returning admins")
	return adminEntities, nil
}

func (cont *OrgController) LoadPublicOrg() error {
	logger.Debug("loading public org")

	orgId := cont.config.Data.Id

	logger.Debugf("reading org with id '%s'", orgId)
	orgPublicJson, err := cont.env.fs.home.Read(orgId)
	if err != nil {
		return err
	}

	logger.Debug("creating new org struct from JSON")
	cont.org, err = entity.New(orgPublicJson)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) SavePublicOrg() error {
	logger.Debug("saving public org")

	logger.Debugf("writing public org with id '%s' to home directory", cont.org.Id())
	if err := cont.env.fs.home.Write(cont.org.Id(), cont.org.DumpPublic()); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) SavePrivateOrg() error {
	logger.Debug("saving private org")

	admins, err := cont.GetOrgAdmins()
	if err != nil {
		return err
	}

	logger.Debug("encrypting org for admins")
	container, err := cont.org.EncryptThenSignString(cont.org.Dump(), admins)
	if err != nil {
		return err
	}

	logger.Debug("sending encrypted org")
	if err := cont.env.api.SendPrivate(cont.org.Id(), cont.org.Id(), container.Dump()); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) LoadPrivateOrg() error {
	logger.Debug("loading private org")

	orgId := cont.config.Data.Id

	logger.Debugf("loading private org with id '%s'", orgId)
	orgEntity, err := cont.env.api.GetPrivate(orgId, orgId)
	if err != nil {
		return err
	}

	logger.Debug("creating new org container")
	contontainer, err := document.NewContainer(orgEntity)
	if err != nil {
		return err
	}

	logger.Debug("verifying container")
	err = cont.org.Verify(contontainer)
	if err != nil {
		return err
	}

	logger.Debug("decrypting container")
	decryptedOrgJson, err := cont.env.controllers.admin.admin.Decrypt(contontainer)
	if err != nil {
		return err
	}

	logger.Debug("creating org struct")
	cont.org, err = entity.New(decryptedOrgJson)
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) CreateIndex() (*index.OrgIndex, error) {
	logger.Debug("creating new org index")

	index, err := index.NewOrg(nil)
	if err != nil {
		return nil, err
	}

	index.Data.Body.Id = NewID()
	index.Data.Body.ParentId = cont.org.Id()

	logger.Trace("returning index")
	return index, nil
}

func (cont *OrgController) GetIndex() (*index.OrgIndex, error) {
	logger.Debug("getting org index")

	orgIndexId := cont.config.Data.Index
	logger.Debugf("getting org index with id '%s'", orgIndexId)
	indexJson, err := cont.env.api.GetPrivate(cont.org.Id(), orgIndexId)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating container for index")
	indexContainer, err := document.NewContainer(indexJson)
	if err != nil {
		return nil, err
	}

	logger.Debug("verifying container")
	err = cont.org.Verify(indexContainer)
	if err != nil {
		return nil, err
	}

	logger.Debug("decrypting container")
	decryptedIndexJson, err := cont.org.Decrypt(indexContainer)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating new index struct from JSON")
	index, err := index.NewOrg(decryptedIndexJson)
	if err != nil {
		return nil, err
	}

	logger.Trace("returning index")
	return index, nil
}

func (cont *OrgController) SaveIndex(index *index.OrgIndex) error {
	logger.Debug("saving org index")
	logger.Tracef("received index with id '%s'", index.Id())

	logger.Debug("encrypting and signing index for org")
	encryptedIndexContainer, err := cont.org.EncryptThenSignString(index.Dump(), nil)
	if err != nil {
		return err
	}

	logger.Debug("sending encrypted index to org")
	err = cont.env.api.SendPrivate(cont.org.Id(), index.Data.Body.Id, encryptedIndexContainer.Dump())
	if err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) GetCA(id string) (*x509.CA, error) {
	logger.Debug("getting CA")
	logger.Tracef("received CA id '%s'", id)

	org := cont.env.controllers.org.org
	caContainerJson, err := cont.env.api.GetPrivate(org.Id(), id)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating CA container")
	caContainer, err := document.NewContainer(caContainerJson)
	if err != nil {
		return nil, err
	}

	logger.Debug("verifying and decrypting CA container")
	caJson, err := org.VerifyThenDecrypt(caContainer)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating new CA from JSON")
	ca, err := x509.NewCA(caJson)
	if err != nil {
		return nil, err
	}

	logger.Trace("returning CA")
	return ca, nil
}

func (cont *OrgController) SignCSR(node *node.Node, caId, tag string) error {
	logger.Debug("signing CSR for node")
	logger.Tracef("received node with id '%s', ca id '%s' and tag '%s'", node.Id(), caId, tag)

	logger.Debugf("popping outgoing CSr from node '%s'", node.Id())
	csrContainerJson, err := cont.env.api.PopOutgoing(node.Data.Body.Id, "csrs")
	if err != nil {
		return err
	}

	logger.Debug("creating new CSR container")
	csrContainer, err := document.NewContainer(csrContainerJson)
	if err != nil {
		return err
	}

	logger.Debug("verifying CSR container with node")
	if err := node.Verify(csrContainer); err != nil {
		return err
	}

	logger.Debug("creating CSR from JSON")
	csrJson := csrContainer.Data.Body
	csr, err := x509.NewCSR(csrJson)
	if err != nil {
		return err
	}

	csr.Data.Body.Name = node.Data.Body.Name

	ca, err := cont.GetCA(caId)
	if err != nil {
		return err
	}

	logger.Debugf("Signing CSR with ca '%s'", caId)
	cert, err := ca.Sign(csr)
	if err != nil {
		return err
	}

	logger.Debug("tagging certificate")
	cert.Data.Body.Tags = append(cert.Data.Body.Tags, tag)

	logger.Debug("creating certificate container")
	certContainer, err := document.NewContainer(nil)
	if err != nil {
		return err
	}

	org := cont.env.controllers.org.org
	certContainer.Data.Options.Source = org.Id()
	certContainer.Data.Body = cert.Dump()

	logger.Debug("signing certificate container with org")
	if err := org.Sign(certContainer); err != nil {
		return err
	}

	logger.Debug("pushing certificate to node")
	if err := cont.env.api.PushIncoming(node.Data.Body.Id, "certs", certContainer.Dump()); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) RegisterNextNode() error {
	logger.Debug("registering next node")

	org := cont.env.controllers.org.org

	logger.Debug("popping next registration from org")
	regJson, err := cont.env.api.PopIncoming(org.Id(), "registration")
	if err != nil {
		return err
	}

	logger.Debug("creating new registration container")
	container, err := document.NewContainer(regJson)
	if err != nil {
		logger.Warn("unable to create container from registration json. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Id(), "registration", regJson)
		return err
	}

	pairingId := container.Data.Options.SignatureInputs["key-id"]
	logger.Debugf("reading pairing key for '%s'", pairingId)

	index, err := cont.GetIndex()
	if err != nil {
		return err
	}

	pairingKey, ok := index.Data.Body.PairingKeys[pairingId]
	if !ok {
		logger.Warn("unable to find pairing key. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Id(), "registration", regJson)
		logger.Trace("returning nil error")
		return nil
	}

	logger.Debug("verifying and decrypting node registration")
	nodeJson, err := org.VerifyAuthenticationThenDecrypt(container, pairingKey.Key)
	if err != nil {
		logger.Warn("unable to decrypt node registration. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Id(), "registration", regJson)
		return err
	}

	logger.Debug("creating new node from JSON")
	node, err := node.New(nodeJson)
	if err != nil {
		logger.Warn("unable to create node. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Id(), "registration", regJson)
		return err
	}

	index.AddEntityTags(node.Data.Body.Id, pairingKey.Tags)
	index.AddNode(node.Data.Body.Name, node.Data.Body.Id)

	logger.Debug("encrypting and signing node for org")
	nodeContainer, err := org.EncryptThenSignString(node.Dump(), nil)
	if err != nil {
		logger.Warn("unable to encrypt node for org. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Id(), "registration", regJson)
		return err
	}

	logger.Debug("sending node to org")
	if err := cont.env.api.SendPrivate(org.Id(), node.Data.Body.Id, nodeContainer.Dump()); err != nil {
		logger.Warn("Unable to send node to org. Pushing back to incoming registration queue")
		cont.env.api.PushIncoming(org.Id(), "registration", regJson)
		return err
	}

	for _, tag := range pairingKey.Tags {
		logger.Debugf("looking for CAs for tag '%s'", tag)
		for _, caId := range index.Data.Body.Tags.CAForward[tag] {
			logger.Debugf("found CA '%s'", caId)
			if err := cont.SignCSR(node, caId, tag); err != nil {
				return err
			}
		}
	}

	if err := cont.SaveIndex(index); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) RegisterNodes() error {
	logger.Debug("registering nodes")

	org := cont.env.controllers.org.org

	for {
		size, err := cont.env.api.IncomingSize(org.Id(), "registration")
		if err != nil {
			return err
		}

		logger.Debugf("found '%d' nodes to register", size)

		if size > 0 {
			if err := cont.RegisterNextNode(); err != nil {
				return err
			}
		} else {
			break
		}
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) Init(params *OrgParams) error {
	logger.Debug("initialising new org")
	logger.Tracef("received params: %s", params)

	if err := params.ValidateOrg(); err != nil {
		return err
	}

	if err := params.ValidateAdmin(); err != nil {
		return err
	}

	if err := cont.env.LoadLocalFs(); err != nil {
		return err
	}

	if err := cont.env.LoadHomeFs(); err != nil {
		return err
	}

	logger.Debugf("checking whether org directory '%s' exists", *params.org)
	exists, err := cont.env.fs.local.Exists(*params.org)
	if err != nil {
		return err
	}

	if exists {
		logger.Warnf("org directory '%s' already exists", *params.org)
		logger.Trace("returning nil error")
		return nil
	}

	cont.env.controllers.admin, err = NewAdminController(cont.env)
	if err != nil {
		return err
	}

	if err := cont.env.controllers.admin.LoadConfig(); err != nil {
		return err
	}

	if cont.env.controllers.admin.config.OrgExists(*params.org) {
		logger.Warnf("org already exists: %s", *params.org)
		logger.Trace("returnin nil error")
		return nil
	}

	logger.Debugf("creating org directory '%s' *params.org")
	if err := cont.env.fs.local.CreateDirectory(*params.org); err != nil {
		return err
	}

	// Make all further fs calls relative to the Org
	logger.Debug("changing to org directory")
	if err := cont.env.fs.local.ChangeToDirectory(*params.org); err != nil {
		return err
	}

	if err := cont.env.LoadAPI(); err != nil {
		return nil
	}

	if err := cont.env.controllers.admin.CreateAdmin(*params.admin); err != nil {
		return err
	}

	admin := cont.env.controllers.admin.admin

	if err := cont.CreateOrg(*params.org); err != nil {
		return err
	}

	if err := cont.LoadConfig(); err != nil {
		return err
	}

	if err := cont.env.controllers.admin.SaveAdmin(); err != nil {
		return err
	}

	orgIndex, err := cont.CreateIndex()
	if err != nil {
		return err
	}

	orgIndex.AddAdmin(admin.Data.Body.Name, admin.Data.Body.Id)

	if err := cont.SaveIndex(orgIndex); err != nil {
		return err
	}

	cont.config.Data.Index = orgIndex.Data.Body.Id
	cont.config.Data.Id = cont.org.Id()
	cont.config.Data.Name = cont.org.Data.Body.Name

	if err := cont.SaveConfig(); err != nil {
		return err
	}

	if err := cont.SavePublicOrg(); err != nil {
		return err
	}

	err = cont.env.controllers.admin.config.AddOrg(cont.org.Data.Body.Name, cont.org.Id(), admin.Data.Body.Id)
	if err != nil {
		return err
	}

	if err := cont.env.controllers.admin.SaveConfig(); err != nil {
		return err
	}

	if err := cont.SavePrivateOrg(); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) List(params *OrgParams) ([]*entity.Entity, error) {
	logger.Debug("listing orgs")
	logger.Tracef("received params: %s", params)

	var err error
	if err := cont.env.LoadHomeFs(); err != nil {
		return nil, err
	}

	if cont.env.controllers.admin == nil {
		cont.env.controllers.admin, err = NewAdminController(cont.env)
		if err != nil {
			return nil, err
		}
	}

	if err := cont.env.controllers.admin.LoadConfig(); err != nil {
		return nil, err
	}

	orgs := make([]*entity.Entity, 0)
	for _, org := range cont.env.controllers.admin.config.GetOrgs() {
		o, err := entity.New(nil)
		if err != nil {
			return nil, err
		}

		o.Data.Body.Id = org.Id
		o.Data.Body.Name = org.Name
		orgs = append(orgs, o)
	}

	logger.Trace("returnings orgs")
	return orgs, nil
}

func (cont *OrgController) Show(params *OrgParams) (*entity.Entity, error) {
	logger.Debug("showing org")
	logger.Tracef("received params: %s", params)

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	logger.Trace("returning org")
	return cont.env.controllers.org.org, nil
}

func (cont *OrgController) RunEnv(params *OrgParams) error {
	logger.Debug("running org tasks")
	logger.Tracef("received params: %s", params)

	if err := cont.RegisterNodes(); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *OrgController) Run(params *OrgParams) error {
	logger.Debug("running org tasks")
	logger.Tracef("received params: %s", params)

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	// LoadAdminEnv has actually loaded a fresh new org controller
	// inside our current env, so all further actions need to be relative
	// to that controller. So we run RunEnv scoped to the correct environment.
	if err := cont.env.controllers.org.RunEnv(params); err != nil {
		return err
	}

	logger.Trace("returning nil err")
	return nil
}

func (cont *OrgController) Delete(params *OrgParams) error {
	logger.Debug("deleting org")
	logger.Tracef("received params: %s", params)

	return fmt.Errorf("Not implemented")
}
