package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
)

func pairingKeyNew(argv map[string]interface{}) (err error) {
	inTags := argv["--tags"].(string)

	app := NewAdminApp()
	app.Load()

	logger.Info("Creating the key")

	id := NewID()
	key := NewID()

	logger.Info("Saving key to index")
	tags := ParseTags(inTags)

	app.LoadOrgIndex()
	// TODO -check error
	app.index.org.AddPairingKey(id, key, tags)
	app.SaveOrgIndex()

	fmt.Printf("Pairing ID: %s\n", id)
	fmt.Printf("Pairing key: %s\n", key)

	return nil
}

func runPairingKey(args []string) (err error) {
	usage := `
Usage:
    pki.io pairing-key [--help]
    pki.io pairing-key new --tags=<tags>

Options:
    --tags=<tags>   Comma-separated list of tags
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["new"].(bool) {
		pairingKeyNew(argv)
	}
	return nil
}
