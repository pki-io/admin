// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
)

func adminCmd(cmd *cli.Cmd) {
	cmd.Command("list", "List admins", adminListCmd)
	cmd.Command("show", "Show an admin", adminShowCmd)
	cmd.Command("invite", "Invite a new admin", adminInviteCmd)
	cmd.Command("new", "Create a new admin", adminNewCmd)
	cmd.Command("run", "Process admin tasks", adminRunCmd)
	cmd.Command("complete", "Complete an admin invite", adminCompleteCmd)
	cmd.Command("delete", "Delete an admin", adminDeleteCmd)
}

func adminListCmd(cmd *cli.Cmd) {
	params := NewAdminParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("listing admins")

		cont, err := NewAdminController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		admins, err := cont.List(params)
		if err != nil {
			app.Fatal(err)
		}

		table := app.NewTable()
		table.SetHeader([]string{"Name", "Id"})

		for _, admin := range admins {
			table.Append([]string{admin.Name(), admin.Id()})
		}

		app.RenderTable(table)
	}
}

func adminShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewAdminParams()
	params.name = cmd.StringArg("NAME", "", "name of admin")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("showing admin")

		cont, err := NewAdminController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		admin, err := cont.Show(params)
		if err != nil {
			app.Fatal(err)
		}

		if admin != nil {

			table := app.NewTable()

			adminData := [][]string{
				[]string{"Name", admin.Name()},
				[]string{"ID", admin.Id()},
			}

			table.AppendBulk(adminData)

			app.RenderTable(table)

			fmt.Println("")
			fmt.Printf("Public signing key:\n%s\n", admin.Data.Body.PublicSigningKey)
			fmt.Printf("Public encryption key:\n%s\n", admin.Data.Body.PublicEncryptionKey)
		}

	}
}

func adminInviteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewAdminParams()
	params.name = cmd.StringArg("NAME", "", "name of admin")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("creating admin invite")

		cont, err := NewAdminController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		keyPair, err := cont.Invite(params)
		if err != nil {
			app.Fatal(err)
		}

		if len(keyPair) > 0 {
			table := app.NewTable()

			adminData := [][]string{
				[]string{"Id", keyPair[0]},
				[]string{"Key", keyPair[1]},
			}

			table.AppendBulk(adminData)

			app.RenderTable(table)
		}

	}
}

func adminNewCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewAdminParams()
	params.name = cmd.StringArg("NAME", "", "name of admin")

	params.inviteId = cmd.StringOpt("invite-id", "", "invite id")
	params.inviteKey = cmd.StringOpt("invite-key", "", "invite key")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("creating new admin")

		cont, err := NewAdminController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.New(params); err != nil {
			app.Fatal(err)
		}

	}
}

func adminRunCmd(cmd *cli.Cmd) {
	params := NewAdminParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("running admin tasks")

		cont, err := NewAdminController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Run(params); err != nil {
			app.Fatal(err)
		}

	}
}

func adminCompleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewAdminParams()
	params.name = cmd.StringArg("NAME", "", "name of admin")

	params.inviteId = cmd.StringOpt("invite-id", "", "invite id")
	params.inviteKey = cmd.StringOpt("invite-key", "", "invite key")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("completing new admin")

		cont, err := NewAdminController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Complete(params); err != nil {
			app.Fatal(err)
		}

	}
}

func adminDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewAdminParams()
	params.name = cmd.StringArg("NAME", "", "name of admin")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting admin")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("deleting admin")

		cont, err := NewAdminController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}

	}
}
