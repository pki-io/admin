package main

import (
	"github.com/jawher/mow.cli"
)

func orgCmd(cmd *cli.Cmd) {
	cmd.Command("list", "List organisations", orgListCmd)
	cmd.Command("show", "Show an organisation", orgShowCmd)
	cmd.Command("run", "Run organisation tasks", orgRunCmd)
	cmd.Command("delete", "Delete an organisation", orgDeleteCmd)
}

func orgListCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := NewOrgParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewOrgController(env)
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

func orgShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := NewOrgParams()
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewOrgController(env)
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

func orgRunCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := NewOrgParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewOrgController(env)
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

func orgDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "ORG [OPTIONS]"

	params := NewOrgParams()
	params.org = cmd.StringArg("ORG", "", "name of organisation")
	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting organisation")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewOrgController(env)
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
