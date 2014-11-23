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
	Example commands:
	foo
	bar
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
	case "admin":
		return runAdmin(argv)
	case "entity":
		return runEntity(argv)
	case "ca":
		return runCA(argv)
	case "client":
		return runClient(argv)
	case "api":
		return runAPI(argv)
	case "help":
		os.Exit(1)
	}
	return fmt.Errorf("%s is not a pki.io command. See 'pki.io help'", cmd)
}

// I understand this monolithic piece of code needs proper breaking up
// but I plan to refactor later ...
func notImpl() (err error) {
	return fmt.Errorf("Not Implemented ...yet")
}

// admin related commands
func runAdmin(argv []string) (err error) {
	return notImpl()
}

// entity related commands
func runEntity(argv []string) (err error) {
	return notImpl()
}

// CA related commands
func runCA(argv []string) (err error) {
	return notImpl()
}

// client related commands
func runClient(argv []string) (err error) {
	return notImpl()
}

// API related commands
func runAPI(argv []string) (err error) {
	return notImpl()
}
