package main

import (
	"fmt"
)

func tagsShow(argv map[string]interface{}) (err error) {

	conf := LoadConfig()
	fsAPI := LoadAPI(conf)
	admin := LoadAdmin(fsAPI)
	org := LoadOrgPrivate(fsAPI, admin)

	fmt.Println("Loading tags")
	tags := LoadTags(fsAPI, org)
	fmt.Println("CA tags:")
	for k, v := range tags.Data.Body.CATags {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("Entity tags:")
	for k, v := range tags.Data.Body.EntityTags {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("Tags to CA:")
	for k, v := range tags.Data.Body.TagCAs {
		fmt.Printf("  %s => %s\n", k, v)
	}
	fmt.Println("Tags to Entity:")
	for k, v := range tags.Data.Body.TagEntities {
		fmt.Printf("  %s => %s\n", k, v)
	}

	return nil
}

// Tags related commands
func runTags(argv map[string]interface{}) (err error) {
	if argv["show"].(bool) {
		tagsShow(argv)
	}
	return nil
}
