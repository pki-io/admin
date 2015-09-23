package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/olekukonko/tablewriter"
	"os"
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
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
		if err != nil {
			env.Fatal(err)
		}

		node, err := cont.New(params)
		if err != nil {
			env.Fatal(err)
		}
		env.logger.Info("Created")

		if node != nil {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"Name", "Id"})

			table.Append([]string{node.Name(), node.Id()})
			env.logger.Flush()
			table.Render()
		}
	}
}

func nodeRunCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
		if err != nil {
			env.Fatal(err)
		}

		env.logger.Info("Running node tasks")
		if err := cont.Run(params); err != nil {
			env.Fatal(err)
		}
		env.logger.Info("Done")
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
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
		if err != nil {
			env.Fatal(err)
		}

		if err := cont.Cert(params); err != nil {
			env.Fatal(err)
		}
	}
}

func nodeListCmd(cmd *cli.Cmd) {
	cmd.Spec = "[OPTIONS]"

	params := NewNodeParams()

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
		if err != nil {
			env.Fatal(err)
		}

		nodes, err := cont.List(params)
		if err != nil {
			env.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeader([]string{"Name", "Id"})

		for _, node := range nodes {
			table.Append([]string{node.Name(), node.Id()})
		}

		env.logger.Flush()
		table.Render()
	}
}

func nodeShowCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := NewNodeParams()
	params.name = cmd.StringArg("NAME", "", "name of node")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
		if err != nil {
			env.Fatal(err)
		}

		env.logger.Info("Showing node")
		node, err := cont.Show(params)
		if err != nil {
			env.Fatal(err)
		}
		env.logger.Info("Done")

		if node != nil {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			nodeData := [][]string{
				[]string{"ID", node.Id()},
				[]string{"Name", node.Name()},
				[]string{"Key type", node.Data.Body.KeyType},
			}

			table.AppendBulk(nodeData)
			env.logger.Flush()
			table.Render()

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
		initLogging(*logLevel, *logging)
		defer logger.Close()
		env := new(Environment)
		env.logger = logger

		cont, err := NewNodeController(env)
		if err != nil {
			env.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			env.Fatal(err)
		}
	}
}
