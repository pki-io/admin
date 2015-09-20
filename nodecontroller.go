package main

import (
	"crypto/x509/pkix"
	"fmt"
	"github.com/pki-io/core/config"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/index"
	"github.com/pki-io/core/node"
	"github.com/pki-io/core/x509"
)

const (
	NodeConfigFile string = "node.conf"
	MinCSRs        int    = 5
)

type NodeParams struct {
	name          *string
	tags          *string
	pairingId     *string
	pairingKey    *string
	confirmDelete *string
	export        *string
	private       *bool
}

func NewNodeParams() *NodeParams {
	return new(NodeParams)
}

func (params *NodeParams) ValidateName(required bool) error {
	if required && *params.name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

func (params *NodeParams) ValidateTags(required bool) error          { return nil }
func (params *NodeParams) ValidateConfirmDelete(required bool) error { return nil }
func (params *NodeParams) ValidateExport(required bool) error        { return nil }
func (params *NodeParams) ValidatePrivate(required bool) error       { return nil }
func (params *NodeParams) ValidatePairingId(required bool) error     { return nil }
func (params *NodeParams) ValidatePairingKey(required bool) error    { return nil }

type NodeController struct {
	env    *Environment
	config *config.NodeConfig
	node   *node.Node
}

func NewNodeController(env *Environment) (*NodeController, error) {
	cont := new(NodeController)
	cont.env = env
	return cont, nil
}

func (cont *NodeController) CreateNode(name string) (*node.Node, error) {
	node, err := node.New(nil)
	if err != nil {
		return nil, err
	}

	node.Data.Body.Name = name
	node.Data.Body.Id = NewID()

	cont.env.logger.Debug("Generating node keys")
	if err := node.GenerateKeys(); err != nil {
		return nil, err
	}

	return node, nil
}

func (cont *NodeController) SecureSendPrivateToOrg(id, key string) error {
	return cont.SecureSendStringToOrg(cont.node.Dump(), id, key)
}

func (cont *NodeController) SecureSendStringToOrg(json, id, key string) error {
	cont.env.logger.Debug("Encrypting data for org")

	org := cont.env.controllers.org.org
	container, err := org.EncryptThenAuthenticateString(json, id, key)
	if err != nil {
		return err
	}

	cont.env.logger.Debug("Pushing container to org")
	if err := cont.env.api.PushIncoming(org.Data.Body.Id, "registration", container.Dump()); err != nil {
		return err
	}

	return nil
}

func (cont *NodeController) CreateIndex() (*index.NodeIndex, error) {

	cont.env.logger.Debug("Creating node index")
	index, err := index.NewNode(nil)
	if err != nil {
		return nil, err
	}

	index.Data.Body.Id = NewID()

	return index, nil
}

func (cont *NodeController) GetIndex() (*index.NodeIndex, error) {
	return nil, nil
}

func (cont *NodeController) SaveIndex(index *index.NodeIndex) error {
	org := cont.env.controllers.org.org

	encryptedIndexContainer, err := org.EncryptThenSignString(index.Dump(), nil)
	if err != nil {
		return err
	}

	if err := cont.env.api.SendPrivate(org.Data.Body.Id, index.Data.Body.Id, encryptedIndexContainer.Dump()); err != nil {
		return err
	}

	return nil
}

func (cont *NodeController) LoadConfig() error {
	var err error

	if cont.config == nil {
		cont.config, err = config.NewNode()
		if err != nil {
			return err
		}
	}

	exists, err := cont.env.fs.local.Exists(NodeConfigFile)
	if err != nil {
		return err
	}

	if exists {
		nodeConfig, err := cont.env.fs.local.Read(NodeConfigFile)
		if err != nil {
			return err
		}

		if err := cont.config.Load(nodeConfig); err != nil {
			return err
		}
	}

	return nil
}

func (cont *NodeController) SaveConfig() error {

	cfgString, err := cont.config.Dump()
	if err != nil {
		return err
	}

	if err := cont.env.fs.local.Write(NodeConfigFile, cfgString); err != nil {
		return err
	}

	return nil
}

func (cont *NodeController) GetNode(name string) (*node.Node, error) {

	org := cont.env.controllers.org.org

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return nil, err
	}

	nodeId, err := index.GetNode(name)
	if err != nil {
		return nil, err
	}

	nodeContainerJson, err := cont.env.api.GetPrivate(org.Data.Body.Id, nodeId)
	if err != nil {
		return nil, err
	}

	nodeContainer, err := document.NewContainer(nodeContainerJson)
	if err != nil {
		return nil, err
	}

	nodeJson, err := org.VerifyThenDecrypt(nodeContainer)
	if err != nil {
		return nil, err
	}

	n, err := node.New(nodeJson)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (cont *NodeController) ProcessNextCert() error {

	certContainerJson, err := cont.env.api.PopIncoming(cont.node.Data.Body.Id, "certs")
	if err != nil {
		return err
	}

	certContainer, err := document.NewContainer(certContainerJson)
	if err != nil {
		return err
	}

	if err := cont.env.controllers.org.org.Verify(certContainer); err != nil {
		return err
	}

	cert, err := x509.NewCertificate(certContainer.Data.Body)
	if err != nil {
		return err
	}

	csrContainerJson, err := cont.env.api.GetPrivate(cont.node.Data.Body.Id, cert.Data.Body.Id)
	if err != nil {
		return err
	}

	csrContainer, err := document.NewContainer(csrContainerJson)
	if err != nil {
		return err
	}

	csrJson, err := cont.node.VerifyThenDecrypt(csrContainer)
	if err != nil {
		return err
	}

	csr, err := x509.NewCSR(csrJson)
	if err != nil {
		return err
	}

	cert.Data.Body.PrivateKey = csr.Data.Body.PrivateKey

	updatedCertContainer, err := cont.node.EncryptThenSignString(cert.Dump(), nil)
	if err != nil {
		return err
	}

	if err := cont.env.api.SendPrivate(cont.node.Data.Body.Id, cert.Data.Body.Id, updatedCertContainer.Dump()); err != nil {
		return err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	if err := index.AddCertTags(cert.Data.Body.Id, cert.Data.Body.Tags); err != nil {
		return err
	}

	if err := cont.env.controllers.org.SaveIndex(index); err != nil {
		return err
	}

	return nil
}

func (cont *NodeController) ProcessCerts() error {

	cont.env.logger.Debug("Processing node certificates")

	for {
		size, err := cont.env.api.IncomingSize(cont.node.Data.Body.Id, "certs")
		if err != nil {
			return err
		}
		cont.env.logger.Debug("Found %d certs to process", size)

		if size > 0 {
			if err := cont.ProcessNextCert(); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (cont *NodeController) CreateCSRs() error {
	cont.env.logger.Debug("Creating CSRs")

	numCSRs, err := cont.env.api.OutgoingSize(cont.node.Data.Body.Id, "csrs")
	if err != nil {
		return err
	}

	for i := 0; i < MinCSRs-numCSRs; i++ {
		if err := cont.NewCSR(); err != nil {
			return err
		}
	}

	return nil
}

func (cont *NodeController) NewCSR() error {
	cont.env.logger.Debug("Creating new CSR")

	csr, err := x509.NewCSR(nil)
	if err != nil {
		return err
	}

	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = cont.node.Data.Body.Name
	subject := pkix.Name{CommonName: csr.Data.Body.Name}
	csr.Generate(&subject)

	cont.env.logger.Debug("Saving local CSR")

	csrContainer, err := cont.node.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		return err
	}

	if err := cont.env.api.SendPrivate(cont.node.Data.Body.Id, csr.Data.Body.Id, csrContainer.Dump()); err != nil {
		return err
	}

	cont.env.logger.Debug("Pushing public CSR")

	csrPublic, err := csr.Public()
	if err != nil {
		return err
	}

	csrPublicContainer, err := cont.node.SignString(csrPublic.Dump())
	if err != nil {
		return err
	}

	if err := cont.env.api.PushOutgoing(cont.node.Data.Body.Id, "csrs", csrPublicContainer.Dump()); err != nil {
		return err
	}

	return nil
}

func (cont *NodeController) New(params *NodeParams) error {
	var err error
	cont.env.logger.Info("Creating new node")

	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.node, err = cont.CreateNode(*params.name)
	if err != nil {
		return err
	}

	if err := cont.SecureSendPrivateToOrg(*params.pairingId, *params.pairingKey); err != nil {
		return err
	}

	index, err := cont.CreateIndex()
	if err != nil {
		return err
	}

	if err := cont.LoadConfig(); err != nil {
		return err
	}

	cont.env.logger.Debug("Adding node to config")
	cont.config.AddNode(cont.node.Data.Body.Name, cont.node.Data.Body.Id, index.Data.Body.Id)

	if err := cont.SaveConfig(); err != nil {
		return err
	}

	if err := cont.CreateCSRs(); err != nil {
		return err
	}

	if err := cont.SaveIndex(index); err != nil {
		return err
	}

	return nil
}

func (cont *NodeController) Run(params *NodeParams) error {

	var err error
	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.node, err = cont.GetNode(*params.name)
	if err != nil {
		return err
	}

	if err := cont.ProcessCerts(); err != nil {
		return err
	}

	return nil
}

func (cont *NodeController) Cert(params *NodeParams) error {

	return fmt.Errorf("Not implemented")
}

func (cont *NodeController) List(params *NodeParams) error {

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return err
	}

	cont.env.logger.Info("Listing nodes:")
	cont.env.logger.Flush()

	for name, id := range index.GetNodes() {
		fmt.Printf("* %s %s\n", name, id)
	}

	return nil
}

func (cont *NodeController) Show(params *NodeParams) error {
	var err error

	cont.env.logger.Debug("Validating parameters")

	if err := params.ValidateName(true); err != nil {
		return err
	}

	cont.env.logger.Debug("Loading admin environment")

	if err := cont.env.LoadAdminEnv(); err != nil {
		return err
	}

	cont.node, err = cont.GetNode(*params.name)
	if err != nil {
		return err
	}

	cont.env.logger.Info("Showing node:")
	cont.env.logger.Flush()

	fmt.Printf("Node name: %s\n", cont.node.Data.Body.Name)
	fmt.Printf("Node ID: %s\n", cont.node.Data.Body.Id)
	fmt.Printf("Public Signing Key:\n%s\n", cont.node.Data.Body.PublicSigningKey)
	fmt.Printf("Public Encryption Key:\n%s\n", cont.node.Data.Body.PublicEncryptionKey)

	return nil
}

func (cont *NodeController) Delete(params *NodeParams) error {

	return fmt.Errorf("Not implemented")
}
