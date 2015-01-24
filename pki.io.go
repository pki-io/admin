package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"os"
)

func main() {
	usage := `
Open source and scalable X.509 certificate management

Usage:
    pki.io [--version] [--help] <command> [<args>...]

Options:
    -h, --help
    -v, --version

Commands:
    init          Initialise an organisation
    ca            Manage X.509 Certificate Authorities
    node          Manage node entities
    org           Do operations on behalf of the org
    pairing-key   Manage pairing keys

See 'pki.io help <command>' for more information on a specific command.
`

	arguments, _ := docopt.Parse(usage, nil, true, "pki.io release 1", true)

	cmd := arguments["<command>"].(string)
	cmdArgs := arguments["<args>"].([]string)

	if "help" == cmd {
		if len(cmdArgs) > 0 {
			cmd = cmdArgs[0]
			cmdArgs = append(cmdArgs, "--help")
		} else {
			fmt.Println(usage)
			os.Exit(0)
		}
	}
	fmt.Printf("Command: %s\n", cmd)
	fmt.Printf("Command args: %s\n", cmdArgs)

	err := runCommand(cmd, cmdArgs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runCommand(cmd string, args []string) error {
	argv := make([]string, 1)
	argv[0] = cmd
	argv = append(argv, args...)
	switch cmd {
	case "init":
		return runInit(argv)
	case "ca":
		return runCA(argv)
	case "node":
		return runNode(argv)
	case "pairing-key":
		return runPairingKey(argv)
	case "org":
		return runOrg(argv)
	}

	return fmt.Errorf("%s is not a pki.io command. See 'pki.io help'", cmd)
}

func notImpl() (err error) {
	return fmt.Errorf("Not Implemented ...yet")
}
