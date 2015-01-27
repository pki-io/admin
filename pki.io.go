package main

import (
	"fmt"
	"errors"
	docopt "github.com/docopt/docopt-go"
	"github.com/cihub/seelog"
)

var logger seelog.LoggerInterface

func init() {
	logger = seelog.Disabled
}

func main() {
	usage := `pki.io
Usage:
  pki.io init <org> [--admin=<admin>] [--logging=<logging>]
  pki.io ca new <name> --tags=<tags> [--parent=<id>] [--logging=<logging>]
  pki.io ca sign <ca> <csr> [--logging=<logging>]
  pki.io csr new <name> [--logging=<logging>]
  pki.io node new <name> --pairing-id=<id> --pairing-key=<key> [--logging=<logging>]
  pki.io node run --name=<name> [--logging=<logging>]
  pki.io node show --name=<name> --cert=<cert> [--logging=<logging>]
  pki.io cert show <name> [--logging=<logging>]
  pki.io org show [--logging=<logging>]
  pki.io org run [--logging=<logging>]
  pki.io pairing-key new --tags=<tags> [--logging=<logging>]
  pki.io --version

Options:
  -h --help      Show this screen
  --version      Show version
  --admin=<name> Administrator name. Defaults to admin.
  --parent=<id>  Parent CA ID
  --tags=<tags>  Comma separated list of tags
  --pairing-key=<key> Pairing key
  --name=<name> Node name
  --cert=<cert> Certificate ID
  --logging=<logging> Logging configuration. Logging is disabled by default.
`
	/*
		Example commands:
		pki.io admin init ENTITY
		pki.io admin revoke ID
		pki.io entity new NAME --offline --parent
		pki.io entity remove ID
		pki.io ca new NAME --tags --parent
		pki.io ca remove ID
		pki.io ca rotate ID
		pki.io ca freeze ID
		pki.io ca revoke ID
		pki.io client new IP --tags
		pki.io client remove ID
		pki.io client rotate ID
		pki.io client freeze ID
		pki.io client revoke ID
		pki.io client revoke ID*/
	arguments, _ := docopt.Parse(usage, nil, true, "pki.io", false)
	initLogging(arguments)
	defer logger.Close()

	logger.Info("test")
	if arguments["init"].(bool) {
		runInit(arguments)
	} else if arguments["org"].(bool) {
		runOrg(arguments)
	} else if arguments["ca"].(bool) {
		runCA(arguments)
	} else if arguments["csr"].(bool) {
		runCSR(arguments)
	} else if arguments["cert"].(bool) {
		runCert(arguments)
	} else if arguments["node"].(bool) {
		runNode(arguments)
	} else if arguments["pairing-key"].(bool) {
		runPairingKey(arguments)
	}
}

// Initialize logging from command arguments.
func initLogging(args map[string]interface {}) {
	if loggingConfig, ok := args["--logging"].(string); ok {
		var err error
		logger, err = seelog.LoggerFromConfigAsFile(loggingConfig)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize logging: %s", err))
		}
	}
}

// I understand this monolithic piece of code needs proper breaking up
// but I plan to refactor later ...
func notImpl() (err error) {
	return errors.New("Not Implemented ...yet")
}
