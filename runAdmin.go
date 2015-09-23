package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/olekukonko/tablewriter"
	"os"
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
		env := new(Environment)
		env.logger = logger

		cont, err := NewAdminController(env)
		if err != nil {
			env.Fatal(err)
		}

		admins, err := cont.List(params)
		if err != nil {
			env.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"Name", "Id"})

		for _, admin := range admins {
			table.Append([]string{admin.Name(), admin.Id()})
		}

		env.logger.Flush()
		table.Render()
	}
}

func adminShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewAdminParams()
	params.name = cmd.StringArg("NAME", "", "name of admin")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewAdminController(env)
		if err != nil {
			env.Fatal(err)
		}

		admin, err := cont.Show(params)
		if err != nil {
			env.Fatal(err)
		}

		if admin != nil {

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			adminData := [][]string{
				[]string{"Name", admin.Name()},
				[]string{"ID", admin.Id()},
			}

			table.AppendBulk(adminData)
			env.logger.Flush()
			table.Render()

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
		env := new(Environment)
		env.logger = logger

		cont, err := NewAdminController(env)
		if err != nil {
			env.Fatal(err)
		}

		keyPair, err := cont.Invite(params)
		if err != nil {
			env.Fatal(err)
		}

		if len(keyPair) > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			adminData := [][]string{
				[]string{"Id", keyPair[0]},
				[]string{"Key", keyPair[1]},
			}

			table.AppendBulk(adminData)
			env.logger.Flush()
			table.Render()
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
		env := new(Environment)
		env.logger = logger

		cont, err := NewAdminController(env)
		if err != nil {
			env.Fatal(err)
		}

		if err := cont.New(params); err != nil {
			env.Fatal(err)
		}

	}
}

func adminRunCmd(cmd *cli.Cmd) {
	params := NewAdminParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewAdminController(env)
		if err != nil {
			env.Fatal(err)
		}

		if err := cont.Run(params); err != nil {
			env.Fatal(err)
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
		env := new(Environment)
		env.logger = logger

		cont, err := NewAdminController(env)
		if err != nil {
			env.Fatal(err)
		}

		if err := cont.Complete(params); err != nil {
			env.Fatal(err)
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
		env := new(Environment)
		env.logger = logger

		cont, err := NewAdminController(env)
		if err != nil {
			env.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			env.Fatal(err)
		}

	}
}
