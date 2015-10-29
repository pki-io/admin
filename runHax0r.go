package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/pki-io/core/document"
	"github.com/pki-io/core/fs"
	"io/ioutil"
)

type Hax0rParams struct {
	jsonFile      *string
	containerFile *string
}

func hax0rCmd(cmd *cli.Cmd) {
	cmd.Command("org-encrypt", "encrypt a json file", orgEncryptCmd)
	cmd.Command("org-decrypt", "decrypt a container", orgDecryptCmd)
}

func orgEncryptCmd(cmd *cli.Cmd) {
	cmd.Spec = "JSON CONTAINER"

	params := new(Hax0rParams)
	params.jsonFile = cmd.StringArg("JSON", "", "json file to encrypt")
	params.containerFile = cmd.StringArg("CONTAINER", "", "container output file")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()

		if err := app.env.LoadAdminEnv(); err != nil {
			app.Fatal(err)
		}

		org := app.env.controllers.org.org

		content, err := fs.ReadFile(*params.jsonFile)
		if err != nil {
			app.Fatal(err)
		}

		container, err := org.EncryptThenSignString(content, nil)
		if err != nil {
			app.Fatal(err)
		}

		if err := ioutil.WriteFile(*params.containerFile, []byte(container.Dump()), 0600); err != nil {
			app.Fatal(err)
		} else {
			fmt.Printf("encrypted container written to %s\n", *params.containerFile)
			fmt.Println("permissions set to 0600. Please fix if requird")
		}

	}
}

func orgDecryptCmd(cmd *cli.Cmd) {
	cmd.Spec = "CONTAINER JSON"

	params := new(Hax0rParams)
	params.containerFile = cmd.StringArg("CONTAINER", "", "container file to decrypt")
	params.jsonFile = cmd.StringArg("JSON", "", "json output file")

	cmd.Action = func() {
		initLogging(*logLevel, *logging)
		defer logger.Close()

		app := NewAdminApp()

		if err := app.env.LoadAdminEnv(); err != nil {
			app.Fatal(err)
		}

		org := app.env.controllers.org.org

		containerJson, err := fs.ReadFile(*params.containerFile)
		if err != nil {
			app.Fatal(err)
		}

		container, err := document.NewContainer(containerJson)
		if err != nil {
			app.Fatal(err)
		}

		json, err := org.VerifyThenDecrypt(container)
		if err != nil {
			app.Fatal(err)
		}

		if err := ioutil.WriteFile(*params.jsonFile, []byte(json), 0600); err != nil {
			app.Fatal(err)
		} else {
			fmt.Printf("decrypted json written to %s\n", *params.jsonFile)
		}
	}
}
