// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/pki-io/controller"
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

	params := controller.NewCertificateParams()
	params.Name = cmd.StringArg("NAME", "", "name of certificate")

	params.Tags = cmd.StringOpt("tags", "NAME", "comma separated list of tags")
	params.StandaloneFile = cmd.StringOpt("standalone", "", "certificate isn't managed by the org but is exported as a tar.gz")
	params.CertFile = cmd.StringOpt("cert", "", "certificate PEM file")
	params.KeyFile = cmd.StringOpt("key", "", "key PEM file")
	params.Expiry = cmd.IntOpt("expiry", 365, "expiry period in days")
	params.Ca = cmd.StringOpt("ca", "", "name of the signing CA (self-signed by default)")
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
		logger.Info("creating new certificate")

		cont, err := controller.NewCertificate(app.env)
		if err != nil {
			app.Fatal(err)
		}

		cert, ca, err := cont.New(params)
		if err != nil {
			app.Fatal(err)
		}

		if cert == nil {
			return
		}

		if *params.StandaloneFile == "" {
			table := app.NewTable()

			certData := [][]string{
				[]string{"Id", cert.Id()},
				[]string{"Name", cert.Name()},
			}
			table.AppendBulk(certData)

			app.RenderTable(table)
		} else {
			var files []ExportFile
			certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
			keyFile := fmt.Sprintf("%s-key.pem", cert.Data.Body.Name)
			caFile := fmt.Sprintf("%s-cacert.pem", cert.Data.Body.Name)

			if ca != nil {
				files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(ca.Data.Body.Certificate)})
			}

			files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})
			files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})

			logger.Debugf("Exporting to '%s'", *params.StandaloneFile)
			Export(files, *params.StandaloneFile)
		}
	}
}

func certListCmd(cmd *cli.Cmd) {
	params := controller.NewCertificateParams()

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("listing certificates")

		cont, err := controller.NewCertificate(app.env)
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

	params := controller.NewCertificateParams()
	params.Name = cmd.StringArg("NAME", "", "name of certificate")

	params.Export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.Private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("showing certificate")

		cont, err := controller.NewCertificate(app.env)
		if err != nil {
			app.Fatal(err)
		}

		cert, err := cont.Show(params)
		if err != nil {
			app.Fatal(err)
		}

		if cert == nil {
			return
		}

		if *params.Export == "" {
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

			if *params.Private {
				fmt.Printf("Private key:\n%s\n", cert.Data.Body.PrivateKey)
			}
		} else {
			var files []ExportFile
			certFile := fmt.Sprintf("%s-cert.pem", cert.Data.Body.Name)
			keyFile := fmt.Sprintf("%s-key.pem", cert.Data.Body.Name)
			files = append(files, ExportFile{Name: certFile, Mode: 0644, Content: []byte(cert.Data.Body.Certificate)})

			if cert.Data.Body.CACertificate != "" {
				caFile := fmt.Sprintf("%s-cacert.pem", cert.Data.Body.Name)
				files = append(files, ExportFile{Name: caFile, Mode: 0644, Content: []byte(cert.Data.Body.CACertificate)})
			}

			if *params.Private {
				files = append(files, ExportFile{Name: keyFile, Mode: 0600, Content: []byte(cert.Data.Body.PrivateKey)})
			}
			logger.Debugf("exporting to '%s'", *params.Export)
			Export(files, *params.Export)
		}

	}
}

func certUpdateCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := controller.NewCertificateParams()
	params.Name = cmd.StringArg("NAME", "", "name of certificate")

	params.CertFile = cmd.StringOpt("cert", "", "certificate PEM file")
	params.KeyFile = cmd.StringOpt("key", "", "key PEM file")
	params.Tags = cmd.StringOpt("tags", "", "comma separated list of tags")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("updating certificate")

		cont, err := controller.NewCertificate(app.env)
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

	params := controller.NewCertificateParams()
	params.Name = cmd.StringArg("NAME", "", "name of certificate")

	params.ConfirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting certificate")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("deleting certificate")

		cont, err := controller.NewCertificate(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}
	}
}
