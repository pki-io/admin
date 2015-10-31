// ThreatSpec package main
package main

import (
	"fmt"
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
		app := NewAdminApp()
		logger.Info("creating new certificate")

		cont, err := NewCertController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		cert, err := cont.New(params)
		if err != nil {
			app.Fatal(err)
		}

		if cert != nil {
			table := app.NewTable()

			certData := [][]string{
				[]string{"Id", cert.Id()},
				[]string{"Name", cert.Name()},
			}
			table.AppendBulk(certData)

			app.RenderTable(table)
		}
	}
}

func certListCmd(cmd *cli.Cmd) {
	params := NewCertParams()

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("listing certificates")

		cont, err := NewCertController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		certs, err := cont.List(params)
		if err != nil {
			app.Fatal(err)
		}

		table := app.NewTable()
		table.SetHeader([]string{"Name", "Id"})

		for _, cert := range certs {
			table.Append([]string{cert.Name(), cert.Id()})
		}

		app.RenderTable(table)
	}
}

func certShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCertParams()
	params.name = cmd.StringArg("NAME", "", "name of certificate")

	params.export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("showing certificate")

		cont, err := NewCertController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		cert, err := cont.Show(params)
		if err != nil {
			app.Fatal(err)
		}

		if cert != nil {

			table := app.NewTable()

			certData := [][]string{
				[]string{"ID", cert.Id()},
				[]string{"Name", cert.Name()},
				[]string{"Key type", cert.Data.Body.KeyType},
			}

			table.AppendBulk(certData)

			app.RenderTable(table)
			fmt.Println("")
			fmt.Printf("Certificate:\n%s\n", cert.Data.Body.Certificate)

			if *params.private {
				fmt.Printf("Private key:\n%s\n", cert.Data.Body.PrivateKey)
			}
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
		app := NewAdminApp()
		logger.Info("updating certificate")

		cont, err := NewCertController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Update(params); err != nil {
			app.Fatal(err)
		}
	}
}

func certDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewCertParams()
	params.name = cmd.StringArg("NAME", "", "name of certificate")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting certificate")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("deleting certificate")

		cont, err := NewCertController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}
	}
}
