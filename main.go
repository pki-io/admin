// ThreatSpec package main
package main

import (
	"github.com/jawher/mow.cli"
	"os"
)

// Global CLI options
var version *bool
var logLevel *string
var logging *string

// ThreatSpec TMv0.1 for main
// Does cli handling for App:CLI
// Receives CLI input from User:CLI to App:CLI
// Calls main.caCmd
func main() {

	cmd := cli.App("pki.io", "Scalable, open source X.509 certificate management")

	// Global options
	logLevel = cmd.StringOpt("l log-level", "info", "log level")
	logging = cmd.StringOpt("logging", "", "alternative logging configuration")

	cmd.Command("init", "Initialize an organization", initCmd)
	cmd.Command("admin", "Manage organization admins", adminCmd)
	cmd.Command("ca", "Manage X.509 Certificate Authorities", caCmd)
	cmd.Command("cert", "Manage X.509 certificates", certCmd)
	cmd.Command("csr", "Manage X.509 Certificate Signing Requests", csrCmd)
	cmd.Command("node", "Manage node entities", nodeCmd)
	cmd.Command("org", "Manage the organization", orgCmd)
	cmd.Command("pairing-key", "Manage pairing keys", pairingKeyCmd)
	cmd.Command("version", "Show version", versionCmd)
	cmd.Command("hax0r", "Leet debug stuff", hax0rCmd)

	cmd.Run(os.Args)
}
