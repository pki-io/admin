package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
)

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
	app.CreateAdminConfig()

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

func runAdmin(args []string) (err error) {
	usage := `
Manages Admins

Usage:
    pki.io admin [--help]
    pki.io admin invite <name>
    pki.io admin new <name> --invite-id=<id> --invite-key=<key>
    pki.io admin run
    pki.io admin complete <name> --invite-id=<id> --invite-key=<key>
Options:
    --invite-id=<id>     Invitation ID
    --invite-key=<key>   Invitation key
`

	argv, _ := docopt.Parse(usage, args, true, "", false)

	if argv["invite"].(bool) {
		adminInvite(argv)
	} else if argv["new"].(bool) {
		adminNew(argv)
	} else if argv["run"].(bool) {
		adminRun(argv)
	} else if argv["complete"].(bool) {
		adminComplete(argv)
	}
	return nil
}
