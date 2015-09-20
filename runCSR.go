package main

import (
	"github.com/jawher/mow.cli"
)

func csrCmd(cmd *cli.Cmd) {
	cmd.Command("new", "Create a new CSR", csrNewCmd)
	cmd.Command("list", "List CSRs", csrListCmd)
	cmd.Command("show", "Show a CSR", csrShowCmd)
	cmd.Command("sign", "Sign a CSR", csrSignCmd)
	cmd.Command("update", "Update a CSR", csrUpdateCmd)
	cmd.Command("delete", "Delete a CSR", csrDeleteCmd)
}

func csrNewCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCSRParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.tags = cmd.StringOpt("tags", "NAME", "comma separated list of tags")
	params.standaloneFile = cmd.StringOpt("standalone", "", "CSR isn't managed by the org but is exported as a tar.gz")
	params.csrFile = cmd.StringOpt("csr", "", "CSR PEM file")
	params.keyFile = cmd.StringOpt("key", "", "key PEM file")
	params.keyType = cmd.StringOpt("key-type", "ec", "Key type (ec or rsa)")
	params.dnLocality = cmd.StringOpt("dn-l", "", "Locality for DN")
	params.dnState = cmd.StringOpt("dn-st", "", "State/province for DN")
	params.dnOrg = cmd.StringOpt("dn-o", "", "Organization for DN")
	params.dnOrgUnit = cmd.StringOpt("dn-ou", "", "Organizational unit for DN")
	params.dnCountry = cmd.StringOpt("dn-c", "", "Country for DN")
	params.dnStreet = cmd.StringOpt("dn-street", "", "Street for DN")
	params.dnPostal = cmd.StringOpt("dn-postal", "", "PostalCode for DN")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCSRController(env)
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

func csrListCmd(cmd *cli.Cmd) {
	params := NewCSRParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCSRController(env)
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

func csrShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCSRParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCSRController(env)
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

func csrSignCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCSRParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.ca = cmd.StringOpt("ca", "", "name of signing CA")
	params.tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCSRController(env)
		if err != nil {
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Sign(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
		}
	}
}
func csrUpdateCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCSRParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.csrFile = cmd.StringOpt("csr", "", "CSR PEM file")
	params.keyFile = cmd.StringOpt("key", "", "key PEM file")
	params.tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCSRController(env)
		if err != nil {
			env.logger.Error(err)
			env.Fatal()
		}

		if err := cont.Update(params); err != nil {
			env.logger.Error(err)
			env.Fatal()
		}
	}
}

func csrDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCSRParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting CSR")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCSRController(env)
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
