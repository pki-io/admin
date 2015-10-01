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

	logger.Debug("Generating node keys")
	if err := node.GenerateKeys(); err != nil {
		return nil, err
	}

	return node, nil
}

func (cont *NodeController) SecureSendPrivateToOrg(id, key string) error {
	return cont.SecureSendStringToOrg(cont.node.Dump(), id, key)
}

func (cont *NodeController) SecureSendStringToOrg(json, id, key string) error {
	logger.Debug("encrypting data for org")
	logger.Tracef("received json [NOT LOGGED] with pairing id '%s' and key [NOT LOGGED]", id)

	org := cont.env.controllers.org.org
	container, err := org.EncryptThenAuthenticateString(json, id, key)
	if err != nil {
		return err
	}

	logger.Debug("pushing container to org with id '%s'", org.Id())
	if err := cont.env.api.PushIncoming(org.Id(), "registration", container.Dump()); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) CreateIndex() (*index.NodeIndex, error) {
	logger.Debug("creating node index")

	index, err := index.NewNode(nil)
	if err != nil {
		return nil, err
	}

	index.Data.Body.Id = NewID()
	logger.Debug("created index with id '%s'", index.Id())

	logger.Trace("returning index")
	return index, nil
}

func (cont *NodeController) SaveIndex(index *index.NodeIndex) error {
	logger.Debug("saving index")
	logger.Tracef("received inded with id '%s'", index.Data.Body.Id)
	org := cont.env.controllers.org.org

	logger.Debug("encrypting and signing index for org")
	encryptedIndexContainer, err := org.EncryptThenSignString(index.Dump(), nil)
	if err != nil {
		return err
	}

	logger.Debug("sending index to org")
	if err := cont.env.api.SendPrivate(org.Id(), index.Data.Body.Id, encryptedIndexContainer.Dump()); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) LoadConfig() error {
	logger.Debug("loading node config")
	var err error

	if cont.config == nil {
		logger.Debug("creating empty config")
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
		logger.Debugf("reading local file '%s'", NodeConfigFile)
		nodeConfig, err := cont.env.fs.local.Read(NodeConfigFile)
		if err != nil {
			return err
		}

		logger.Debug("loading config")
		if err := cont.config.Load(nodeConfig); err != nil {
			return err
		}
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) SaveConfig() error {
	logger.Debug("saving node config")

	logger.Debug("dumping config")
	cfgString, err := cont.config.Dump()
	if err != nil {
		return err
	}

	logger.Debugf("writing config to local file '%s'", NodeConfigFile)
	if err := cont.env.fs.local.Write(NodeConfigFile, cfgString); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) GetNode(name string) (*node.Node, error) {
	logger.Debug("getting node")
	logger.Tracef("received name '%s'", name)

	org := cont.env.controllers.org.org

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return nil, err
	}

	nodeId, err := index.GetNode(name)
	if err != nil {
		return nil, err
	}

	logger.Debugf("getting node '%s' from org", nodeId)
	nodeContainerJson, err := cont.env.api.GetPrivate(org.Id(), nodeId)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating new node container")
	nodeContainer, err := document.NewContainer(nodeContainerJson)
	if err != nil {
		return nil, err
	}

	logger.Debug("verifying and decrypting node container")
	nodeJson, err := org.VerifyThenDecrypt(nodeContainer)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating new node struct")
	n, err := node.New(nodeJson)
	if err != nil {
		return nil, err
	}

	logger.Trace("returning node")
	return n, nil
}

func (cont *NodeController) ProcessNextCert() error {
	logger.Debug("processing next certificate")

	logger.Debug("getting next incoming certificate JSON")
	certContainerJson, err := cont.env.api.PopIncoming(cont.node.Data.Body.Id, "certs")
	if err != nil {
		return err
	}

	logger.Debug("creating certificate container from JSON")
	certContainer, err := document.NewContainer(certContainerJson)
	if err != nil {
		return err
	}

	logger.Debug("verifying container is signed by org")
	if err := cont.env.controllers.org.org.Verify(certContainer); err != nil {
		return err
	}

	logger.Debug("creating new certificate struct")
	cert, err := x509.NewCertificate(certContainer.Data.Body)
	if err != nil {
		return err
	}

	logger.Debugf("getting matching CSR for id '%s'", cert.Data.Body.Id)
	csrContainerJson, err := cont.env.api.GetPrivate(cont.node.Data.Body.Id, cert.Data.Body.Id)
	if err != nil {
		return err
	}

	logger.Debug("creating CSR container")
	csrContainer, err := document.NewContainer(csrContainerJson)
	if err != nil {
		return err
	}

	logger.Debug("verifying and decryping CSR container")
	csrJson, err := cont.node.VerifyThenDecrypt(csrContainer)
	if err != nil {
		return err
	}

	logger.Debug("creating CSR struct from JSON")
	csr, err := x509.NewCSR(csrJson)
	if err != nil {
		return err
	}

	logger.Debug("setting certificate private key from CSR")
	cert.Data.Body.PrivateKey = csr.Data.Body.PrivateKey

	logger.Debug("encrypting and signing certificate for node")
	updatedCertContainer, err := cont.node.EncryptThenSignString(cert.Dump(), nil)
	if err != nil {
		return err
	}

	logger.Debug("saving encrypted certificate for node")
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

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) ProcessCerts() error {
	logger.Debug("processing node certificates")

	for {
		logger.Debug("getting number of incoming certificates")
		size, err := cont.env.api.IncomingSize(cont.node.Data.Body.Id, "certs")
		if err != nil {
			return err
		}
		logger.Debugf("found %d certificates to process", size)

		if size > 0 {
			if err := cont.ProcessNextCert(); err != nil {
				return err
			}
		} else {
			break
		}
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) CreateCSRs() error {
	logger.Debug("creating CSRs")

	logger.Debug("getting number of outgoing CSRs")
	numCSRs, err := cont.env.api.OutgoingSize(cont.node.Data.Body.Id, "csrs")
	if err != nil {
		return err
	}
	logger.Debugf("found '%d' CSRs", numCSRs)

	for i := 0; i < MinCSRs-numCSRs; i++ {
		if err := cont.NewCSR(); err != nil {
			return err
		}
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) NewCSR() error {
	logger.Debug("creating new CSR")

	csr, err := x509.NewCSR(nil)
	if err != nil {
		return err
	}

	csr.Data.Body.Id = NewID()
	csr.Data.Body.Name = cont.node.Data.Body.Name
	subject := pkix.Name{CommonName: csr.Data.Body.Name}
	csr.Generate(&subject)

	logger.Debug("creating encrypted CSR container")
	csrContainer, err := cont.node.EncryptThenSignString(csr.Dump(), nil)
	if err != nil {
		return err
	}

	logger.Debug("saving node CSR")
	if err := cont.env.api.SendPrivate(cont.node.Data.Body.Id, csr.Data.Body.Id, csrContainer.Dump()); err != nil {
		return err
	}

	logger.Debug("getting public CSR")
	csrPublic, err := csr.Public()
	if err != nil {
		return err
	}

	logger.Debug("signing public CSR as node")
	csrPublicContainer, err := cont.node.SignString(csrPublic.Dump())
	if err != nil {
		return err
	}

	logger.Debug("putting public CSR in outgoing queue")
	if err := cont.env.api.PushOutgoing(cont.node.Data.Body.Id, "csrs", csrPublicContainer.Dump()); err != nil {
		return err
	}

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) New(params *NodeParams) (*node.Node, error) {
	logger.Debug("creating new new")
	logger.Tracef("received params: %s", params)
	var err error

	if err := params.ValidateName(true); err != nil {
		return nil, err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	cont.node, err = cont.CreateNode(*params.name)
	if err != nil {
		return nil, err
	}

	logger.Debugf("sending registration to org with pairing id '%s'", *params.pairingId)
	if err := cont.SecureSendPrivateToOrg(*params.pairingId, *params.pairingKey); err != nil {
		return nil, err
	}

	index, err := cont.CreateIndex()
	if err != nil {
		return nil, err
	}

	if err := cont.LoadConfig(); err != nil {
		return nil, err
	}

	cont.config.AddNode(cont.node.Data.Body.Name, cont.node.Data.Body.Id, index.Data.Body.Id)

	if err := cont.SaveConfig(); err != nil {
		return nil, err
	}

	if err := cont.CreateCSRs(); err != nil {
		return nil, err
	}

	if err := cont.SaveIndex(index); err != nil {
		return nil, err
	}

	logger.Trace("returning node")
	return cont.node, nil
}

func (cont *NodeController) Run(params *NodeParams) error {
	logger.Debug("running node tasks")
	logger.Tracef("received params: %s", params)

	var err error

	if err := params.ValidateName(true); err != nil {
		return err
	}

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

	logger.Trace("returning nil error")
	return nil
}

func (cont *NodeController) Cert(params *NodeParams) error {
	logger.Debug("getting certificates for node")
	logger.Tracef("received params: %s", params)
	return fmt.Errorf("not implemented")
}

func (cont *NodeController) List(params *NodeParams) ([]*node.Node, error) {
	logger.Debug("listing nodes")
	logger.Tracef("received params: %s", params)

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	index, err := cont.env.controllers.org.GetIndex()
	if err != nil {
		return nil, err
	}

	nodes := make([]*node.Node, 0)
	for name, _ := range index.GetNodes() {
		node, err := cont.GetNode(name)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	logger.Trace("returning nodes")
	return nodes, nil
}

func (cont *NodeController) Show(params *NodeParams) (*node.Node, error) {
	logger.Debug("showing node")
	logger.Tracef("received params: %s", params)
	var err error

	if err := params.ValidateName(true); err != nil {
		return nil, err
	}

	if err := cont.env.LoadAdminEnv(); err != nil {
		return nil, err
	}

	cont.node, err = cont.GetNode(*params.name)
	if err != nil {
		return nil, err
	}

	logger.Trace("returning node")
	return cont.node, nil
}

func (cont *NodeController) Delete(params *NodeParams) error {
	logger.Debug("deleting node")
	logger.Tracef("received params: %s", params)
	return fmt.Errorf("Not implemented")
}
