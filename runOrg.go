package main

import (
	"fmt"
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

	return nil
}

func runOrg(argv map[string]interface{}) (err error) {
	if argv["show"].(bool) {
		orgShow(argv)
	}
	return nil
}
