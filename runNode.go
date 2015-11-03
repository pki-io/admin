// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
)

func nodeCmd(cmd *cli.Cmd) {
	cmd.Command("new", "Create a new node", nodeNewCmd)
	cmd.Command("run", "Run tasks for node", nodeRunCmd)
	cmd.Command("cert", "Get certificates for node", nodeCertCmd)
	cmd.Command("list", "List nodes", nodeListCmd)
	cmd.Command("show", "Show a node", nodeShowCmd)
	cmd.Command("delete", "Delete a node", nodeDeleteCmd)
}

func nodeNewCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	params.pairingId = cmd.StringOpt("pairing-id", "", "pairing id")
	params.pairingKey = cmd.StringOpt("pairing-key", "", "pairing key")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("creating new node")

		cont, err := NewNodeController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		node, err := cont.New(params)
		if err != nil {
			app.Fatal(err)
		}

		if node != nil {
			table := app.NewTable()
			table.SetHeader([]string{"Name", "Id"})

			table.Append([]string{node.Name(), node.Id()})

			app.RenderTable(table)
		}
	}
}

func nodeRunCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("running node tasks")

		cont, err := NewNodeController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Run(params); err != nil {
			app.Fatal(err)
		}
	}
}

func nodeCertCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")
	params.tags = cmd.StringOpt("tags", "NAME", "comma separated list of tags")
	params.export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("creating node certificate")

		cont, err := NewNodeController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Cert(params); err != nil {
			app.Fatal(err)
		}
	}
}

func nodeListCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := NewNodeParams()

	cmd.Action = func() {
		app := NewAdminApp()

		logger.Info("listing nodes")

		cont, err := NewNodeController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		nodes, err := cont.List(params)
		if err != nil {
			app.Fatal(err)
		}

		table := app.NewTable()
		table.SetHeader([]string{"Name", "Id"})

		for _, node := range nodes {
			table.Append([]string{node.Name(), node.Id()})
		}

		app.RenderTable(table)
		app.Exit()
	}
}

func nodeShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("showing node")

		cont, err := NewNodeController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		node, err := cont.Show(params)
		if err != nil {
			app.Fatal(err)
		}

		if node != nil {
			table := app.NewTable()

			nodeData := [][]string{
				[]string{"ID", node.Id()},
				[]string{"Name", node.Name()},
				[]string{"Key type", node.Data.Body.KeyType},
			}
			table.AppendBulk(nodeData)

			app.RenderTable(table)
			fmt.Println("")
			fmt.Printf("Public signing key:\n%s\n", node.Data.Body.PublicSigningKey)
			fmt.Printf("Public encryption key:\n%s\n", node.Data.Body.PublicEncryptionKey)
		}
	}
}

func nodeDeleteCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	params.confirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting node")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("deleting node")
		cont, err := NewNodeController(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}
	}
}
