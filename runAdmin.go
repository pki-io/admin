package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
)

func adminList(argv map[string]interface{}) (err error) {
	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()
	admins, err := app.index.org.GetAdmins()
	checkAppFatal("Unable to get admins: %s", err)

	logger.Info("Admins:")
	// Flush?
	for name, id := range admins {
		fmt.Printf("* %s %s\n", name, id)
	}
	return nil
}

func adminShow(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)

	app := NewAdminApp()
	app.Load()
	app.LoadOrgIndex()
	adminId, err := app.index.org.GetAdmin(name)
	checkAppFatal("Unable to get admin id: %s", err)

	admin := app.GetAdminEntity(adminId)

	fmt.Printf("Name: %s\n", admin.Data.Body.Name)
	fmt.Printf("ID: %s\n", admin.Data.Body.Id)
	fmt.Printf("Key type: %s\n", admin.Data.Body.KeyType)
	fmt.Printf("Public encryption key:\n%s\n", admin.Data.Body.PublicEncryptionKey)
	fmt.Printf("Public signing key:\n%s\n", admin.Data.Body.PublicSigningKey)
	return nil
}

func adminInvite(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)

	app := NewAdminApp()
	app.Load()

	logger.Info("Creating the key")
	id := NewID()
	key := NewID()

	logger.Info("Saving key to index")
	app.LoadOrgIndex()
	app.index.org.AddInviteKey(id, key, name)
	app.SaveOrgIndex()

	fmt.Printf("Invite ID: %s\n", id)
	fmt.Printf("Invite key: %s\n", key)

	return nil
}

func adminNew(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	inviteId := ArgString(argv["--invite-id"], nil)
	inviteKey := ArgString(argv["--invite-key"], nil)

	app := NewAdminApp()

	app.InitLocalFs()
	app.LoadOrgConfig()
	app.InitHomeFs()
	app.InitApiFs()

	app.CreateAdminEntity(name)

	exists, err := app.AdminConfigExists()
	checkAppFatal("Could not check admin config existence: %s", err)

	if exists {
		logger.Info("Existing admin config found")
		app.LoadAdminConfig()
	} else {
		app.CreateAdminConfig()
	}
	app.config.admin.AddOrg(app.config.org.Data.Name, app.config.org.Data.Id, app.entities.admin.Data.Body.Id)

	app.SaveAdminEntity()
	app.SaveAdminConfig()

	app.SecureSendPublicToOrg(inviteId, inviteKey)
	return nil
}

func adminRun(argv map[string]interface{}) (err error) {

	app := NewAdminApp()
	app.Load()

	app.LoadOrgIndex()
	app.ProcessInvites()
	app.SaveOrgIndex()
	return nil
}

func adminComplete(argv map[string]interface{}) (err error) {
	//name := ArgString(argv["<name>"], nil)
	inviteId := ArgString(argv["--invite-id"], nil)
	inviteKey := ArgString(argv["--invite-key"], nil)

	app := NewAdminApp()

	app.InitLocalFs()
	app.LoadOrgConfig()
	app.InitHomeFs()
	app.LoadAdminConfig()
	app.InitApiFs()
	app.LoadAdminEntity()

	app.CompleteInvite(inviteId, inviteKey)

	return nil
}

func adminDelete(argv map[string]interface{}) (err error) {
	name := ArgString(argv["<name>"], nil)
	reason := ArgString(argv["--confirm-delete"], nil)

	logger.Infof("Deleting admin %s: %s", name, reason)
	app := NewAdminApp()
	app.Load()

	app.LoadOrgIndex()
	err = app.index.org.RemoveAdmin(name)
	checkAppFatal("Can't delete admin: %s", err)

	app.SaveOrgIndex()
	app.SendOrgEntity()
	return nil
}

func runAdmin(args []string) (err error) {
	usage := `
Manages Admins

Usage:
    pki.io admin [--help]
    pki.io admin list
    pki.io admin show <name>
    pki.io admin invite <name>
    pki.io admin new <name> --invite-id <id> --invite-key <key>
    pki.io admin run
    pki.io admin complete <name> --invite-id <id> --invite-key <key>
    pki.io admin delete <name> --confirm-delete <reason>
Options:
    --invite-id <id>                Invitation ID
    --invite-key <key>              Invitation key
    --confirm-delete <reason>       Reason for deleting admin
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["invite"].(bool) {
		adminInvite(argv)
	} else if argv["list"].(bool) {
		adminList(argv)
	} else if argv["show"].(bool) {
		adminShow(argv)
	} else if argv["new"].(bool) {
		adminNew(argv)
	} else if argv["run"].(bool) {
		adminRun(argv)
	} else if argv["complete"].(bool) {
		adminComplete(argv)
	} else if argv["delete"].(bool) {
		adminDelete(argv)
	}
	return nil
}
