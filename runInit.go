package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/pki-io/pki.io/config"
	"github.com/pki-io/pki.io/entity"
	"github.com/pki-io/pki.io/fs"
	"github.com/pki-io/pki.io/index"
	"os"
	"path/filepath"
)

func runInit(args []string) (err error) {

	usage := `Usage: pki.io init <org> [--admin=<admin>]

Options
    --admin   Admin name. Defaults to 'admin'
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	var adminName string
	orgName := argv["<org>"].(string)
	if argv["--admin"] == nil {
		adminName = "admin"
	} else {
		adminName = argv["--admin"].(string)
	}

	/**************************************************************************************************
	 * Initialise the file system
	 **************************************************************************************************/

	currentDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Could not get current directory: %s", err.Error()))
	}

	// This will create the directory for us $(cwd)/NAME
	fsAPI, err := fs.NewAPI(currentDir, orgName)
	if err != nil {
		panic(fmt.Sprintf("Could not initialise the filesystem API: %S", err.Error()))
	}

	/**************************************************************************************************
	 * Create the Org
	 **************************************************************************************************/

	fmt.Println("Creating Org entity")
	org, err := entity.New(nil)
	if err != nil {
		panic(fmt.Sprintf("Could not create org entity: %s", err.Error()))
	}

	// Need an ID (perhaps it should be the API via a register call?)
	org.Data.Body.Id = NewID()
	org.Data.Body.Name = orgName

	fmt.Println("Generating Org keys")
	err = org.GenerateKeys()
	if err != nil {
		panic(fmt.Sprintf("Could not generate org keys: %s", err.Error()))
	}

	// Public view of the Org
	fmt.Println("Creating public copy of org to save locally")
	publicOrg, err := org.Public()
	if err != nil {
		panic(fmt.Sprintf("Could not get public org: %s", err.Error()))
	}

	// Index
	fmt.Println("Creating org index")
	index, err := index.New(nil)
	if err != nil {
		panic(fmt.Sprintf("Could not create index: %s", err.Error()))
	}

	/**************************************************************************************************
	 * Create the admin
	 **************************************************************************************************/

	fmt.Println("Creating Admin entity")
	admin, err := entity.New(nil)
	if err != nil {
		panic(fmt.Sprintf("Could not create admin entity: %s", err.Error()))
	}

	// Need an ID (perhaps it should be the API via a register call?)
	admin.Data.Body.Id = NewID()
	admin.Data.Body.Name = adminName

	// Generate admin keys
	err = admin.GenerateKeys()
	if err != nil {
		panic(fmt.Sprintf("Could not generate admin keys: %s", err.Error()))
	}

	// API is for our admin user (should be a login call?)
	fsAPI.Id = admin.Data.Body.Id

	/**************************************************************************************************
	 * Create the config file
	 **************************************************************************************************/

	// Save the org config
	configFile := filepath.Join(fsAPI.Path, "pki.io.conf")
	conf := config.New(configFile)

	conf.AddOrg(org.Data.Body.Name, org.Data.Body.Id)
	conf.AddAdmin(admin.Data.Body.Name, admin.Data.Body.Id)

	if err := conf.Create(); err != nil {
		panic(fmt.Sprintf("Could not create org config: %s", err.Error()))
	}

	/**************************************************************************************************
	 * Save admin to API
	 **************************************************************************************************/

	fmt.Println("Saving admin")
	if err := fsAPI.WriteLocal("admin", admin.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save admin data: %s", err.Error()))
	}

	/**************************************************************************************************
	 * Save the org to API
	 **************************************************************************************************/

	// Public keys
	publicOrgContainer, err := admin.SignString(publicOrg.Dump())
	if err != nil {
		panic(fmt.Sprintf("Could not sign public org: %s", err.Error()))
	}

	fmt.Println("Saving org data")
	if err := fsAPI.StorePublic("org", publicOrgContainer.Dump()); err != nil {
		panic(fmt.Sprintf("Could not save file: %s", err.Error()))
	}

	// Private keys
	container, err := admin.EncryptThenSignString(org.Dump(), nil)
	if err != nil {
		panic(fmt.Sprintf("Could encrypt org: %s", err.Error()))
	}

	if err := fsAPI.StorePrivate("org", container.Dump()); err != nil {
		panic(fmt.Sprintf("Could not store container to json: %s", err.Error()))
	}

	// Index
	SaveIndex(fsAPI, org, index)

	return nil

}
