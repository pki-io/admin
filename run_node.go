// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/pki-io/controller"
	"github.com/pki-io/core/ssh"
	"os"
	"strings"
)

func nodeCmd(cmd *cli.Cmd) {
	cmd.Command("init", "Initialise agent on remote node", nodeInitCmd)
	cmd.Command("new", "Create a new node", nodeNewCmd)
	cmd.Command("run", "Run tasks for node", nodeRunCmd)
	cmd.Command("cert", "Get certificates for node", nodeCertCmd)
	cmd.Command("list", "List nodes", nodeListCmd)
	cmd.Command("show", "Show a node", nodeShowCmd)
	cmd.Command("delete", "Delete a node", nodeDeleteCmd)
}

func nodeInitCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME HOST [OPTIONS] [-- SSH_ARGS]"

	name := cmd.StringArg("NAME", "", "name of node")
	host := cmd.StringArg("HOST", "", "node hostname or ip address")

	agentFile := cmd.StringOpt("agent-file", "", "path to agent package")
	installFile := cmd.StringOpt("install-file", "./agent-installer.sh", "path to agent installer script")

	sshArgs := cmd.StringArg("SSH_ARGS", "", "arguments to pass to ssh NOT WORKING YET")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Infof("initialising node %s", *name)

		s, err := ssh.Connect(*host, strings.Split(*sshArgs, " "))
		if err != nil {
			app.Fatal(err)
		}

		if err := s.PutFiles("", *agentFile); err != nil {
			app.Fatal(err)
		}

		if err := s.PutFiles("", *installFile); err != nil {
			app.Fatal(err)
		}

		// stream output in real time
		if err := s.Execute("sh agent-installer.sh", nil, os.Stdout, os.Stderr); err != nil {
			app.Fatal(err)
		}

		if err := s.Close(); err != nil {
			app.Fatal(err)
		}
	}
}

func nodeNewCmd(cmd *cli.Cmd) {
	cmd.Spec = "NAME [OPTIONS]"

	params := controller.NewNodeParams()
	params.Name = cmd.StringArg("NAME", "", "name of node")

	params.PairingId = cmd.StringOpt("pairing-id", "", "pairing id")
	params.PairingKey = cmd.StringOpt("pairing-key", "", "pairing key")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("creating new node")

		cont, err := controller.NewNode(app.env)
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

	params := controller.NewNodeParams()
	params.Name = cmd.StringArg("NAME", "", "name of node")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("running node tasks")

		cont, err := controller.NewNode(app.env)
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

	params := controller.NewNodeParams()
	params.Name = cmd.StringArg("NAME", "", "name of node")
	params.Tags = cmd.StringOpt("tags", "NAME", "comma separated list of tags")
	params.Export = cmd.StringOpt("export", "", "tar.gz export to file")
	params.Private = cmd.BoolOpt("private", false, "show/export private data")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("creating node certificate")

		cont, err := controller.NewNode(app.env)
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

	params := controller.NewNodeParams()

	cmd.Action = func() {
		app := NewAdminApp()

		logger.Info("listing nodes")

		cont, err := controller.NewNode(app.env)
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

	params := controller.NewNodeParams()
	params.Name = cmd.StringArg("NAME", "", "name of node")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("showing node")

		cont, err := controller.NewNode(app.env)
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

	params := controller.NewNodeParams()
	params.Name = cmd.StringArg("NAME", "", "name of node")

	params.ConfirmDelete = cmd.StringOpt("confirm-delete", "", "reason for deleting node")

	cmd.Action = func() {
		app := NewAdminApp()
		logger.Info("deleting node")
		cont, err := controller.NewNode(app.env)
		if err != nil {
			app.Fatal(err)
		}

		if err := cont.Delete(params); err != nil {
			app.Fatal(err)
		}
	}
}
