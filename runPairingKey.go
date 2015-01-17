package main

import (
	"encoding/hex"
	"fmt"
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
	key := hex.EncodeToString(crypto.RandomBytes(16))
	tags := ParseTags(inTags)
	index.AddPairingKey(id, key, tags)

	SaveIndex(fsAPI, org, index)
	fmt.Printf("Pairing ID: %s\n", id)
	fmt.Printf("Pairing key: %s\n", key)
	return nil
}

func runPairingKey(argv map[string]interface{}) (err error) {
	if argv["new"].(bool) {
		pairingKeyNew(argv)
	}
	return nil
}
