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
		env := new(Environment)
		env.logger = logger

		cont, err := NewPairingKeyController(env)
		if err != nil {
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.New(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.List(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
		}
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Show(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Delete(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
		}
	}
}
