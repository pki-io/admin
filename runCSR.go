// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/pki-io/controllers/csr"
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

	params := csr.NewParams()
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

		app := NewAdminApp()
		logger.Info("creating new CSR")

		cont, err := csr.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		csr, err := cont.New(params)
		if err != nil {
			app.Fatal(err)
		}

		if csr != nil {
			table := app.NewTable()

			csrData := [][]string{
				[]string{"Id", csr.Id()},
				[]string{"Name", csr.Name()},
			}
			table.AppendBulk(csrData)

			app.RenderTable(table)
			fmt.Println("")
			fmt.Printf("CSR:\n%s\n", csr.Data.Body.CSR)
		}
	}
}

func csrListCmd(cmd *cli.Cmd) {
	params := csr.NewParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("listing CSRs")

		cont, err := csr.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		csrs, err := cont.List(params)
		if err != nil {
			app.Fatal(err)
		}

		table := app.NewTable()
		table.SetHeader([]string{"Name", "Id"})

		for _, csr := range csrs {
			table.Append([]string{csr.Name(), csr.Id()})
		}

		app.RenderTable(table)
	}
}

func csrShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := csr.NewParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("showing CSR")

		cont, err := csr.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		csr, err := cont.Show(params)
		if err != nil {
			app.Fatal(err)
		}

		if csr != nil {
			table := app.NewTable()

			csrData := [][]string{
				[]string{"Id", csr.Id()},
				[]string{"Name", csr.Name()},
				[]string{"Key type", csr.Data.Body.KeyType},
			}

			table.AppendBulk(csrData)

			app.RenderTable(table)

			fmt.Printf("CSR:\n%s\n", csr.Data.Body.CSR)
			if *params.private {
				fmt.Printf("Private key:\n%s\n", csr.Data.Body.PrivateKey)
			}
		}
	}
}

func csrSignCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := csr.NewParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.keepSubject = cmd.BoolOpt("keep-subject", false, "keep subject from CSR")
	params.ca = cmd.StringOpt("ca", "", "name of signing CA")
	params.tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("signing CSR")

		cont, err := csr.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		cert, err := cont.Sign(params)
		if err != nil {
			app.Fatal(err)
		}

		if cert != nil {
			table := app.NewTable()

			certData := [][]string{
				[]string{"Id", cert.Id()},
				[]string{"Name", cert.Name()},
				[]string{"Key type", cert.Data.Body.KeyType},
			}

			table.AppendBulk(certData)

			app.RenderTable(table)
			fmt.Printf("Certificate:\n%s\n", cert.Data.Body.Certificate)
		}
	}
}

func csrUpdateCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := csr.NewParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.csrFile = cmd.StringOpt("csr", "", "CSR PEM file")
	params.keyFile = cmd.StringOpt("key", "", "key PEM file")
	params.tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("updating CSR")

		cont, err := csr.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Update(params); err != nil {
			app.Fatal(err)
		}
	}
}

func csrDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := csr.NewParams()
	params.name = cmd.StringArg("NAME", "", "name of CSR")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting CSR")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()
		logger.Info("deleting CSR")

		cont, err := csr.New(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}
	}
}
