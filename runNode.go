package main

import (
	"github.com/jawher/mow.cli"
)

func nodeCmd(cmd *cli.Cmd) {
	cmd.Command("new", "Create a new node", nodeNewCmd)
	cmd.Command("run", "Run tasks for node", nodeRunCmd)
	cmd.Command("cert", "Get certificates for node", nodeCertCmd)
	cmd.Command("list", "List nodes", nodeListCmd)
	cmd.Command("show", "Show a node", nodeShowCmd)
	cmd.Command("delete", "Delete a node", nodeDeleteCmd)
}

func nodeNewCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	params.pairingId = cmd.StringOpt("pairing-id", "", "pairing id")
	params.pairingKey = cmd.StringOpt("pairing-key", "", "pairing key")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
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

func nodeRunCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
		if err != nil {
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Run(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
		}
	}
}

func nodeCertCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")
	params.tags = cmd.StringOpt("tags", "NAME", "comma separated list of tags")
	params.export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
		if err != nil {
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Cert(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
		}
	}
}

func nodeListCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := NewNodeParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
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

func nodeShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
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

func nodeDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting node")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
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
