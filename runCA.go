package main

import (
	"github.com/jawher/mow.cli"
)

func caCmd(cmd *cli.Cmd) {
	cmd.Command("new", "Create a new CA", caNewCmd)
	cmd.Command("list", "List CAs", caListCmd)
	cmd.Command("show", "Show a CA", caShowCmd)
	cmd.Command("update", "Update an existing CA", caUpdateCmd)
	cmd.Command("delete", "Delete a CA", caDeleteCmd)
}

func caNewCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCAParams()
	params.name = cmd.StringArg("NAME", "", "name of CA")

	params.certFile = cmd.StringOpt("cert", "", "certificate PEM file")
	params.keyFile = cmd.StringOpt("key", "", "key PEM file")
	params.tags = cmd.StringOpt("tags", "NAME", "comma separated list of tags")
	params.caExpiry = cmd.IntOpt("ca-expiry", 365, "CA expiry period in days")
	params.certExpiry = cmd.IntOpt("cert-expiry", 90, "Certificate expiry period in days")
	params.keyType = cmd.StringOpt("key-type", "ec", "Key type (ec or rsa)")
	params.dnLocality = cmd.StringOpt("dn-l", "", "Locality for DN scope")
	params.dnState = cmd.StringOpt("dn-st", "", "State/province for DN scope")
	params.dnOrg = cmd.StringOpt("dn-o", "", "Organization for DN scope")
	params.dnOrgUnit = cmd.StringOpt("dn-ou", "", "Organizational unit for DN scope")
	params.dnCountry = cmd.StringOpt("dn-c", "", "Country for DN scope")
	params.dnStreet = cmd.StringOpt("dn-street", "", "Street for DN scope")
	params.dnPostal = cmd.StringOpt("dn-postal", "", "PostalCode for DN scope")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCAController(env)
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

func caListCmd(cmd *cli.Cmd) {
	params := NewCAParams()

	cmd.Action = func() {

		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCAController(env)
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

func caShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCAParams()
	params.name = cmd.StringArg("NAME", "", "name of CA")

	params.export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCAController(env)
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

func caUpdateCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCAParams()
	params.name = cmd.StringArg("NAME", "", "name of CA")

	params.certFile = cmd.StringOpt("cert", "", "certificate PEM file")
	params.keyFile = cmd.StringOpt("key", "", "key PEM file")
	params.tags = cmd.StringOpt("tags", "", "comma separated list of tags")
	params.caExpiry = cmd.IntOpt("ca-expiry", 0, "CA expiry period in days")
	params.certExpiry = cmd.IntOpt("cert-expiry", 0, "Certificate expiry period in days")
	params.dnLocality = cmd.StringOpt("dn-l", "", "Locality for DN scope")
	params.dnState = cmd.StringOpt("dn-st", "", "State/province for DN scope")
	params.dnOrg = cmd.StringOpt("dn-o", "", "Organization for DN scope")
	params.dnOrgUnit = cmd.StringOpt("dn-ou", "", "Organizational unit for DN scope")
	params.dnCountry = cmd.StringOpt("dn-c", "", "Country for DN scope")
	params.dnStreet = cmd.StringOpt("dn-street", "", "Street for DN scope")
	params.dnPostal = cmd.StringOpt("dn-postal", "", "PostalCode for DN scope")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCAController(env)
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

func caDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCAParams()
	params.name = cmd.StringArg("NAME", "", "name of CA")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting CA")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewCAController(env)
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
