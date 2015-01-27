package main

import (
	"fmt"
	"errors"
	docopt "github.com/docopt/docopt-go"
)

func main() {
	usage := `pki.io
Usage:
  pki.io init <org> [--admin=<admin>]
  pki.io ca new <name> --tags=<tags> [--parent=<id>] 
  pki.io ca sign <ca> <csr>
  pki.io csr new <name>
  pki.io node new <name> --pairing-id=<id> --pairing-key=<key>
  pki.io node run --name=<name>
  pki.io node show --name=<name> --cert=<cert>
  pki.io cert show <name>
  pki.io org show
  pki.io org run
  pki.io pairing-key new --tags=<tags>
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
	fmt.Println(arguments)

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

// I understand this monolithic piece of code needs proper breaking up
// but I plan to refactor later ...
func notImpl() (err error) {
	return errors.New("Not Implemented ...yet")
}
