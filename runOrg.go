package main

import (
	"fmt"
)

func orgShow(argv map[string]interface{}) (err error) {

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPublic(fsAPI, admin)

	fmt.Printf("Name: %s\n", org.Data.Body.Name)
	fmt.Printf("Id: %s\n", org.Data.Body.Id)
	fmt.Printf("Public Signing Key:\n%s\n", org.Data.Body.PublicSigningKey)
	fmt.Printf("Public Encryption Key:\n%s\n", org.Data.Body.PublicEncryptionKey)

	return nil
}

func runOrg(argv map[string]interface{}) (err error) {
	if argv["show"].(bool) {
		orgShow(argv)
	}
	return nil
}
