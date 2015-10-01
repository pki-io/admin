package main

import (
	"github.com/jawher/mow.cli"
)

func pairingKeyCmd(cmd *cli.Cmd) {
	cmd.Command("new", "Create a new pairing key", pairingKeyNewCmd)
	cmd.Command("list", "List pairing keys", pairingKeyListCmd)
	cmd.Command("show", "Show a pairing key", pairingKeyShowCmd)
	cmd.Command("delete", "Delete a pairing key", pairingKeyDeleteCmd)
}

func pairingKeyNewCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := NewPairingKeyParams()
	params.tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("creating new pairing key")

		cont, err := NewPairingKeyController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		id, key, err := cont.New(params)
		if err != nil {
			app.Fatal(err)
		}

		if id != "" && key != "" {
			table := app.NewTable()

			keyData := [][]string{
				[]string{"Id", id},
				[]string{"Key", key},
			}
			table.AppendBulk(keyData)

			app.RenderTable(table)
		}
	}

}

func pairingKeyListCmd(cmd *cli.Cmd) {
	params := NewPairingKeyParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("listing pairing keys")

		cont, err := NewPairingKeyController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		keys, err := cont.List(params)
		if err != nil {
			app.Fatal(err)
		}

		table := app.NewTable()
		table.SetHeader([]string{"Id", "Tags"})

		for _, key := range keys {
			table.Append([]string{key[0], key[1]})
		}

		app.RenderTable(table)
	}
}

func pairingKeyShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "ID [OPTIONS]"

	params := NewPairingKeyParams()
	params.id = cmd.StringArg("ID", "", "Public ID of pairing key")

	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("showing pairing key")

		cont, err := NewPairingKeyController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		id, key, tags, err := cont.Show(params)
		if err != nil {
			app.Fatal(err)
		}

		if id != "" && key != "" && tags != "" {
			table := app.NewTable()

			table.Append([]string{"Id", id})
			if *params.private {
				table.Append([]string{"Key", key})
			}
			table.Append([]string{"Tags", tags})

			app.RenderTable(table)
		}
	}
}

func pairingKeyDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "ID [OPTIONS]"

	params := NewPairingKeyParams()
	params.id = cmd.StringArg("ID", "", "Public ID of pairing key")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting pairing key")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("deleting pairing key")

		cont, err := NewPairingKeyController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}
	}
}
