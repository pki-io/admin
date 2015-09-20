package main

import (
	"github.com/jawher/mow.cli"
)

func certCmd(cmd *cli.Cmd) {
	cmd.Command("new", "Create a new certificate", certNewCmd)
	cmd.Command("list", "List certificates", certListCmd)
	cmd.Command("show", "Show a certificate", certShowCmd)
	cmd.Command("update", "Update a certificate", certUpdateCmd)
	cmd.Command("delete", "Delete a certificate", certDeleteCmd)
}

func certNewCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCertParams()
	params.name = cmd.StringArg("NAME", "", "name of certificate")

	params.tags = cmd.StringOpt("tags", "NAME", "comma separated list of tags")
	params.standaloneFile = cmd.StringOpt("standalone", "", "certificate isn't managed by the org but is exported as a tar.gz")
	params.certFile = cmd.StringOpt("cert", "", "certificate PEM file")
	params.keyFile = cmd.StringOpt("key", "", "key PEM file")
	params.expiry = cmd.IntOpt("expiry", 365, "expiry period in days")
	params.ca = cmd.StringOpt("ca", "", "name of the signing CA (self-signed by default)")
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

		cont, err := NewCertController(env)
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

func certListCmd(cmd *cli.Cmd) {
	params := NewCertParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCertController(env)
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

func certShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCertParams()
	params.name = cmd.StringArg("NAME", "", "name of certificate")

	params.export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCertController(env)
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

func certUpdateCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCertParams()
	params.name = cmd.StringArg("NAME", "", "name of certificate")

	params.certFile = cmd.StringOpt("cert", "", "certificate PEM file")
	params.keyFile = cmd.StringOpt("key", "", "key PEM file")
	params.tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCertController(env)
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

func certDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCertParams()
	params.name = cmd.StringArg("NAME", "", "name of certificate")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting certificate")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCertController(env)
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
