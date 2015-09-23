package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/olekukonko/tablewriter"
	"os"
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
			env.Fatal(err)
		}

		orgs, err := cont.List(params)
		if err != nil {
			env.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"Name", "Id"})

		for _, org := range orgs {
			table.Append([]string{org.Name(), org.Id()})
		}

		env.logger.Flush()
		table.Render()
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
			env.Fatal(err)
		}

		org, err := cont.Show(params)
		if err != nil {
			env.Fatal(err)
		}

		if org != nil {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			orgData := [][]string{
				[]string{"ID", org.Id()},
				[]string{"Name", org.Name()},
				[]string{"Key type", org.Data.Body.KeyType},
			}

			table.AppendBulk(orgData)
			env.logger.Flush()
			table.Render()

			fmt.Println("")
			fmt.Printf("Public signing key:\n%s\n", org.Data.Body.PublicSigningKey)
			fmt.Printf("Public encryption key:\n%s\n", org.Data.Body.PublicEncryptionKey)
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
			env.Fatal(err)
		}

		if err := cont.Run(params); err != nil {
			env.Fatal(err)
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
			env.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			env.Fatal(err)
		}
	}
}
