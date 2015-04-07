package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
)

func nodeNew(argv map[string]interface{}) (err error) {
	name := argv["<name>"].(string)
	pairingId := argv["--pairing-id"].(string)
	pairingKey := argv["--pairing-key"].(string)

	adminApp := NewAdminApp()
	adminApp.Load()

	nodeApp := NewNodeApp()
	nodeApp.InitLocalFs()

	nodeApp.InitApiFs()
	nodeApp.InitHomeFs()

	nodeApp.entities.org = adminApp.entities.org

	nodeApp.CreateNodeEntity(name)

	nodeApp.SecureSendPrivateToOrg(pairingId, pairingKey)

	logger.Info("Switching to node context")
	nodeApp.fs.api.Authenticate(nodeApp.entities.node.Data.Body.Id, "")

	nodeApp.CreateNodeIndex()
	nodeApp.CreateNodeConfig()
	nodeApp.SaveNodeConfig()

	logger.Info("Creating CSRs")
	nodeApp.GenerateCSRs()

	nodeApp.SaveNodeIndex()

	return nil
}

func nodeRun(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)

	adminApp := NewAdminApp()
	adminApp.Load()

	nodeApp := NewNodeApp()
	nodeApp.Load()

	nodeApp.entities.org = adminApp.entities.org

	adminApp.fs.api.Authenticate(adminApp.entities.org.Data.Body.Id, "")

	adminApp.LoadOrgIndex()

	nodeApp.entities.node = adminApp.GetNode(name)
	nodeApp.fs.api.Authenticate(nodeApp.entities.node.Data.Body.Id, "")

	nodeApp.LoadNodeIndex()
	nodeApp.ProcessCerts()
	nodeApp.SaveNodeIndex()
	return nil
}

func nodeShow(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)

	adminApp := NewAdminApp()
	adminApp.Load()

	nodeApp := NewNodeApp()
	nodeApp.Load()

	nodeApp.entities.org = adminApp.entities.org

	adminApp.fs.api.Authenticate(adminApp.entities.org.Data.Body.Id, "")

	adminApp.LoadOrgIndex()

	nodeApp.entities.node = adminApp.GetNode(name)
	nodeApp.fs.api.Authenticate(nodeApp.entities.node.Data.Body.Id, "")

	nodeApp.LoadNodeIndex()

	logger.Infof("Node name: %s", nodeApp.entities.node.Data.Body.Name)
	logger.Infof("Node ID: %s", nodeApp.entities.node.Data.Body.Id)
	logger.Infof("Public Signing Key:\n%s", nodeApp.entities.node.Data.Body.PublicSigningKey)
	logger.Infof("Public Encryption Key:\n%s", nodeApp.entities.node.Data.Body.PublicEncryptionKey)
	logger.Infof("Certificate tags:\n%s", nodeApp.index.node.Data.Body.Tags.CertForward)

	return nil
}

func nodeCert(argv map[string]interface{}) (err error) {
	name := argv["--name"].(string)
	inTags := argv["--tags"].(string)
	exportFile := argv["--export"]

	adminApp := NewAdminApp()
	adminApp.Load()

	nodeApp := NewNodeApp()
	nodeApp.Load()

	nodeApp.entities.org = adminApp.entities.org

	adminApp.fs.api.Authenticate(adminApp.entities.org.Data.Body.Id, "")

	adminApp.LoadOrgIndex()

	nodeApp.entities.node = adminApp.GetNode(name)
	nodeApp.fs.api.Authenticate(nodeApp.entities.node.Data.Body.Id, "")

	nodeApp.LoadNodeIndex()

	certs := nodeApp.GetCertificates(inTags)

	var files []ExportFile
	for _, cert := range certs {
		switch exportFile.(type) {
		case nil:
			logger.Infof("Subject: %s", cert.Data.Body.Name)
			logger.Infof("Certificate:\n%s", cert.Data.Body.Certificate)
			logger.Infof("Private Key:\n%s", cert.Data.Body.PrivateKey)
			logger.Infof("CA Certificate:\n%s", cert.Data.Body.CACertificate)
		case string:
			certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
			keyFile := fmt.Sprintf("%s-key.pem", cert.Data.Body.Name)
			caFile := fmt.Sprintf("%s-cacert.pem", cert.Data.Body.Name)
			files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})
			files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(cert.Data.Body.CACertificate)})
		}
	}

	if len(files) > 0 {
		logger.Info("Exporting")
		Export(files, exportFile.(string))
	}

	return nil
}

func nodeList(argv map[string]interface{}) (err error) {
	adminApp := NewAdminApp()
	adminApp.Load()

	adminApp.LoadOrgIndex()
	logger.Info("Nodes:")
	for name, id := range adminApp.index.org.GetNodes() {
		fmt.Printf("* %s %s\n", name, id)
	}
	return nil
}

func nodeDelete(argv map[string]interface{}) (err error) {
	name := ArgString(argv["--name"], nil)
	reason := ArgString(argv["--confirm-delete"], nil)

	adminApp := NewAdminApp()
	adminApp.Load()
	adminApp.fs.api.Authenticate(adminApp.entities.org.Data.Body.Id, "")

	nodeApp := NewNodeApp()
	nodeApp.Load()

	nodeApp.entities.org = adminApp.entities.org

	adminApp.LoadOrgIndex()
	nodeApp.entities.node = adminApp.GetNode(name)
	nodeApp.fs.api.Authenticate(nodeApp.entities.node.Data.Body.Id, "")
	nodeApp.LoadNodeIndex()

	logger.Infof("Deleting node %s with reason %s", name, reason)
	logger.Info("Deleting node index")
	err = nodeApp.fs.api.DeletePrivate(nodeApp.index.node.Data.Body.Id)
	checkAppFatal("Could not delete node index: %s", err)

	logger.Info("Removing node from config")
	err = nodeApp.config.node.RemoveNode(name)
	checkAppFatal("Could not remove node from config: %s", err)
	nodeApp.SaveNodeConfig()

	logger.Info("Removing node from org index")
	adminApp.LoadOrgIndex()
	err = adminApp.index.org.RemoveNode(name)
	checkAppFatal("Could not remove node from org index", err)
	adminApp.SaveOrgIndex()

	return nil
}

func runNode(args []string) (err error) {

	usage := `
Manages nodes.

Usage:
    pki.io node [--help]
    pki.io node new <name> --pairing-id=<id> --pairing-key=<key> [--offline]
    pki.io node run --name=<name>
    pki.io node show --name=<name>
    pki.io node cert --name=<name> --tags=<tags> [--export=<file>]
    pki.io node list
    pki.io node delete --name=<name> --confirm-delete=<reason>


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
		nodeRun(argv)
	} else if argv["show"].(bool) {
		nodeShow(argv)
	} else if argv["cert"].(bool) {
		nodeCert(argv)
	} else if argv["list"].(bool) {
		nodeList(argv)
	} else if argv["delete"].(bool) {
		nodeDelete(argv)
	}
	return nil
}
