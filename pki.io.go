package main

import (
	"fmt"
	"os"
	// http://docopt.org/
	"github.com/docopt/docopt-go"
)

func main() {
	usage := `pki.io
Usage:
    pki.io [--version] <command> [<args>...]
Options:
    -h --help  Show this screen
    -v --version  Show version
    `

	arguments, _ := docopt.Parse(usage, nil, true, "pki.io", false)
	fmt.Println(arguments)

	fmt.Println("command arguments:")
	cmd := arguments["<command>"].(string)
	cmdArgs := arguments["<args>"].([]string)
	fmt.Println(cmdArgs)
	err := runCommand(cmd, cmdArgs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runCommand(cmd string, args []string) (err error) {
	argv := make([]string, 1)
	argv[0] = cmd
	argv = append(argv, args...)
	fmt.Println(cmd)
	switch cmd {
	case "admin", "entity", "ca", "client", "api", "help":
		return fmt.Errorf("Not Implemented yet")
	}
	return fmt.Errorf("%s is not a pki.io command. See 'pki.io help'", cmd)
	return
}
