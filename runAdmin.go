package main

import (
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
		env := new(Environment)
		env.logger = logger

		cont, err := NewAdminController(env)
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Show(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Invite(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.New(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Run(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Complete(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
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
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Delete(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
		}

	}
}
