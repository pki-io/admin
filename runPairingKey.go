package main

import (
	"github.com/jawher/mow.cli"
	"github.com/olekukonko/tablewriter"
	"os"
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
		env := new(Environment)
		env.logger = logger

		cont, err := NewPairingKeyController(env)
		if err != nil {
			env.Fatal(err)
		}

		id, key, err := cont.New(params)
		if err != nil {
			env.Fatal(err)
		}

		if id != "" && key != "" {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			keyData := [][]string{
				[]string{"Id", id},
				[]string{"Key", key},
			}

			table.AppendBulk(keyData)
			env.logger.Flush()
			table.Render()
		}
	}

}

func pairingKeyListCmd(cmd *cli.Cmd) {
	params := NewPairingKeyParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewPairingKeyController(env)
		if err != nil {
			env.Fatal(err)
		}

		keys, err := cont.List(params)
		if err != nil {
			env.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"Id", "Tags"})
		for _, key := range keys {
			table.Append([]string{key[0], key[1]})
		}

		env.logger.Flush()
		table.Render()
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
		env := new(Environment)
		env.logger = logger

		cont, err := NewPairingKeyController(env)
		if err != nil {
			env.Fatal(err)
		}

		id, key, tags, err := cont.Show(params)
		if err != nil {
			env.Fatal(err)
		}

		if id != "" && key != "" && tags != "" {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			table.Append([]string{"Id", id})
			if *params.private {
				table.Append([]string{"Key", key})
			}
			table.Append([]string{"Tags", tags})

			env.logger.Flush()
			table.Render()
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
		env := new(Environment)
		env.logger = logger

		cont, err := NewPairingKeyController(env)
		if err != nil {
			env.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			env.Fatal(err)
		}
	}
}
