// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/pki-io/controllers/org"
)

func orgCmd(cmd *cli.Cmd) {
	cmd.Command("list", "List organisations", orgListCmd)
	cmd.Command("show", "Show an organisation", orgShowCmd)
	cmd.Command("run", "Run organisation tasks", orgRunCmd)
	cmd.Command("delete", "Delete an organisation", orgDeleteCmd)
}

func orgListCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := org.NewParams()

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("listing organisations")

		cont, err := org.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		orgs, err := cont.List(params)
		if err != nil {
			app.Fatal(err)
		}

		table := app.NewTable()
		table.SetHeader([]string{"Name", "Id"})

		for _, org := range orgs {
			table.Append([]string{org.Name(), org.Id()})
		}

		app.RenderTable(table)
	}
}

func orgShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := org.NewParams()
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("showing organisation")

		cont, err := org.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		org, err := cont.Show(params)
		if err != nil {
			app.Fatal(err)
		}

		if org != nil {
			table := app.NewTable()

			orgData := [][]string{
				[]string{"ID", org.Id()},
				[]string{"Name", org.Name()},
				[]string{"Key type", org.Data.Body.KeyType},
			}
			table.AppendBulk(orgData)

			app.RenderTable(table)
			fmt.Println("")
			fmt.Printf("Public signing key:\n%s\n", org.Data.Body.PublicSigningKey)
			fmt.Printf("Public encryption key:\n%s\n", org.Data.Body.PublicEncryptionKey)
		}
	}
}

func orgRunCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := org.NewParams()

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("running organisation tasks")

		cont, err := org.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Run(params); err != nil {
			app.Fatal(err)
		}
	}
}

func orgDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "ORG [OPTIONS]"

	params := org.NewParams()
	params.org = cmd.StringArg("ORG", "", "name of organisation")
	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting organisation")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("deleting organisation")

		cont, err := org.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}
	}
}
