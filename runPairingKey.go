package main

import (
	"encoding/hex"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/pki.io/crypto"
)

func pairingKeyNew(argv map[string]interface{}) (err error) {
	inTags := argv["--tags"].(string)

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)
	index := LoadIndex(fsAPI, org)

	id := NewID()
	random, _ := crypto.RandomBytes(16)
	key := hex.EncodeToString(random)
	tags := ParseTags(inTags)
	index.AddPairingKey(id, key, tags)

	SaveIndex(fsAPI, org, index)
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
