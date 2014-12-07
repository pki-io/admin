package main

import (
	"fmt"
	//"os"
	// http://docopt.org/
	"github.com/docopt/docopt-go"
)

func main() {
	usage := `pki.io
Usage:
  pki.io init --org=<org> --admin=<admin>
  pki.io org show
  pki.io --version

Options:
  -h --help   Show this screen
  --version   Show version
  --org=<name> Organisation name
  --admin=<name> Administrator name
`
/*
	Example commands:
	pki.io admin init ENTITY
	pki.io admin revoke ID
	pki.io entity new NAME --offline --parent
	pki.io entity remove ID
	pki.io ca new NAME --tags --parent
	pki.io ca remove ID
	pki.io ca rotate ID
	pki.io ca freeze ID
	pki.io ca revoke ID
	pki.io client new IP --tags
	pki.io client remove ID
	pki.io client rotate ID
	pki.io client freeze ID
	pki.io client revoke ID
	pki.io client revoke ID*/
	arguments, _ := docopt.Parse(usage, nil, true, "pki.io", false)
	fmt.Println(arguments)

        if arguments["init"].(bool) {
            runInit(arguments)
        } else if arguments["org"].(bool) {
            runOrg(arguments)
        }
	//fmt.Println("command arguments:")
	//cmd := arguments["<command>"].(string)
	//cmdArgs := arguments["<args>"].([]string)
	//fmt.Println(cmdArgs)
	//err := runCommand(cmd, cmdArgs)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
}


// I understand this monolithic piece of code needs proper breaking up
// but I plan to refactor later ...
func notImpl() (err error) {
	return fmt.Errorf("Not Implemented ...yet")
}
