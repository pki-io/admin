package main

import (
	"fmt"
	"pki.io/crypto"
	"pki.io/node"
)

func orgShow(argv map[string]interface{}) (err error) {

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)
	index := LoadIndex(fsAPI, org)

	fmt.Printf("Name: %s\n", org.Data.Body.Name)
	fmt.Printf("Id: %s\n", org.Data.Body.Id)
	fmt.Printf("Public Signing Key:\n%s\n", org.Data.Body.PublicSigningKey)
	fmt.Printf("Public Encryption Key:\n%s\n", org.Data.Body.PublicEncryptionKey)

	fmt.Println("Tags for CAs:")
	for k, v := range index.Data.Body.Tags.CAForward {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("CA tags:")
	for k, v := range index.Data.Body.Tags.CAReverse {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("Tags for entities:")
	for k, v := range index.Data.Body.Tags.EntityForward {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("Entity tags:")
	for k, v := range index.Data.Body.Tags.EntityReverse {
		fmt.Printf("  %s => %s\n", k, v)
	}

	fmt.Println("Pairing keys:")
	for k, v := range index.Data.Body.PairingKeys {
		fmt.Printf("  %s => %s\n", k, v)
	}
	return nil
}

func orgRegisterNodes(argv map[string]interface{}) (err error) {
	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)
	indx := LoadIndex(fsAPI, org)

	for {
		size, err := fsAPI.IncomingSize("registration")
		if err != nil {
			panic(fmt.Sprintf("Can't get queue size: %s", err.Error()))
		}
		if size > 0 {
			regJson, err := fsAPI.PopIncoming("registration")
			if err != nil {
				panic(fmt.Sprintf("Can't pop registration: %s", err.Error()))
			}

			nodeReg, err := node.NewRegistration(regJson)
			if err != nil {
				panic(fmt.Sprintf("Can't load registration: %s", err.Error()))
			}

			pairingId := nodeReg.Data.Options.PairingId
			pairingKey := indx.Data.Body.PairingKeys[pairingId]

			macKey, _ := crypto.ExpandKey([]byte(pairingKey.Key), []byte(nodeReg.Data.Options.SignatureSalt))
			if err := nodeReg.Verify(string(macKey)); err != nil {
				panic(fmt.Sprintf("Couldn't verify registration: %s", err.Error()))
			}

			node, err := node.NewFromRegistration(nodeReg)
			if err != nil {
				panic(fmt.Sprintf("Couldn't create node from registration: %s", err.Error()))
			}
			indx.AddEntityTags(node.Data.Body.Id, pairingKey.Tags)
			// register node in some way
		} else {
			break
		}
	}

	SaveIndex(fsAPI, org, indx)
	return nil
}

func runOrg(argv map[string]interface{}) (err error) {
	if argv["show"].(bool) {
		orgShow(argv)
	} else if argv["register-nodes"].(bool) {
		orgRegisterNodes(argv)
	}
	return nil
}
