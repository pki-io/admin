// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/pki-io/controller"
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

	params := controller.NewCSRParams()
	params.Name = cmd.StringArg("NAME", "", "name of CSR")

	params.Tags = cmd.StringOpt("tags", "NAME", "comma separated list of tags")
	params.StandaloneFile = cmd.StringOpt("standalone", "", "CSR isn't managed by the org but is exported as a tar.gz")
	params.CsrFile = cmd.StringOpt("csr", "", "CSR PEM file")
	params.KeyFile = cmd.StringOpt("key", "", "key PEM file")
	params.KeyType = cmd.StringOpt("key-type", "ec", "Key type (ec or rsa)")
	params.DnLocality = cmd.StringOpt("dn-l", "", "Locality for DN")
	params.DnState = cmd.StringOpt("dn-st", "", "State/province for DN")
	params.DnOrg = cmd.StringOpt("dn-o", "", "Organization for DN")
	params.DnOrgUnit = cmd.StringOpt("dn-ou", "", "Organizational unit for DN")
	params.DnCountry = cmd.StringOpt("dn-c", "", "Country for DN")
	params.DnStreet = cmd.StringOpt("dn-street", "", "Street for DN")
	params.DnPostal = cmd.StringOpt("dn-postal", "", "PostalCode for DN")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("creating new CSR")

		cont, err := controller.NewCSR(app.env)
		if err != nil {
			app.Fatal(err)
		}

		csr, err := cont.New(params)
		if err != nil {
			app.Fatal(err)
		}

		if csr == nil {
			return
		}

		if *params.StandaloneFile == "" {
			table := app.NewTable()

			csrData := [][]string{
				[]string{"Id", csr.Id()},
				[]string{"Name", csr.Name()},
			}
			table.AppendBulk(csrData)

			app.RenderTable(table)
			fmt.Println("")
			fmt.Printf("CSR:\n%s\n", csr.Data.Body.CSR)
		} else {
			var files []ExportFile
			csrFile := fmt.Sprintf("%s-csr.pem", csr.Data.Body.Name)
			keyFile := fmt.Sprintf("%s-key.pem", csr.Data.Body.Name)

			files = append(files, ExportFile{Name: csrFile, Mode: 0644, Content: []byte(csr.Data.Body.CSR)})
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(csr.Data.Body.PrivateKey)})

			logger.Debugf("exporting to '%s'", *params.StandaloneFile)
			Export(files, *params.StandaloneFile)
		}
	}
}

func csrListCmd(cmd *cli.Cmd) {
	params := controller.NewCSRParams()

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("listing CSRs")

		cont, err := controller.NewCSR(app.env)
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

	params := controller.NewCSRParams()
	params.Name = cmd.StringArg("NAME", "", "name of CSR")

	params.Export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.Private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("showing CSR")

		cont, err := controller.NewCSR(app.env)
		if err != nil {
			app.Fatal(err)
		}

		csr, err := cont.Show(params)
		if err != nil {
			app.Fatal(err)
		}

		if csr == nil {
			return
		}

		if *params.Export == "" {
			table := app.NewTable()

			csrData := [][]string{
				[]string{"Id", csr.Id()},
				[]string{"Name", csr.Name()},
				[]string{"Key type", csr.Data.Body.KeyType},
			}

			table.AppendBulk(csrData)

			app.RenderTable(table)

			fmt.Printf("CSR:\n%s\n", csr.Data.Body.CSR)
			if *params.Private {
				fmt.Printf("Private key:\n%s\n", csr.Data.Body.PrivateKey)
			}
		} else {
			var files []ExportFile
			csrFile := fmt.Sprintf("%s-csr.pem", csr.Data.Body.Name)
			keyFile := fmt.Sprintf("%s-key.pem", csr.Data.Body.Name)

			files = append(files, ExportFile{Name: csrFile, Mode: 0644, Content: []byte(csr.Data.Body.CSR)})

			if *params.Private {
				files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(csr.Data.Body.PrivateKey)})
			}

			logger.Infof("Exporting to '%s'", *params.Export)
			Export(files, *params.Export)
		}
	}
}

func csrSignCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := controller.NewCSRParams()
	params.Name = cmd.StringArg("NAME", "", "name of CSR")

	params.KeepSubject = cmd.BoolOpt("keep-subject", false, "keep subject from CSR")
	params.Ca = cmd.StringOpt("ca", "", "name of signing CA")
	params.Tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("signing CSR")

		cont, err := controller.NewCSR(app.env)
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

	params := controller.NewCSRParams()
	params.Name = cmd.StringArg("NAME", "", "name of CSR")

	params.CsrFile = cmd.StringOpt("csr", "", "CSR PEM file")
	params.KeyFile = cmd.StringOpt("key", "", "key PEM file")
	params.Tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("updating CSR")

		cont, err := controller.NewCSR(app.env)
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

	params := controller.NewCSRParams()
	params.Name = cmd.StringArg("NAME", "", "name of CSR")

	params.ConfirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting CSR")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("deleting CSR")

		cont, err := controller.NewCSR(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}
	}
}
