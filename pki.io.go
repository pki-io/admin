package main

import (
	"errors"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/docopt/docopt-go"
	"os"
)

const defaultLoggingConfig string = `
<seelog minlevel="info" maxlevel="error">
	<outputs formatid="raw">
		<console/>
	</outputs>
	<formats>
		<format id="raw" format="%Msg%n"/>
	</formats>
</seelog>
`

var logger seelog.LoggerInterface

func init() {
	var err error
	logger, err = seelog.LoggerFromConfigAsString(defaultLoggingConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to load default logging configuration.\n%s", err))
	}
}

func main() {
	usage := `
Open source and scalable X.509 certificate management

Usage:
    pki.io [--version] [--help] [--logging=<logging>] <command> [<args>...]

Options:
    -h, --help
    -v, --version
    --logging=<logging> Logging configuration. Logging is disabled by default.

Commands:
    init          Initialise an organisation
    ca            Manage X.509 Certificate Authorities
    node          Manage node entities
    org           Do operations on behalf of the org
    pairing-key   Manage pairing keys

See 'pki.io help <command>' for more information on a specific command.
`

	arguments, _ := docopt.Parse(usage, nil, true, "pki.io release 1", true)

	initLogging(arguments)
	defer logger.Close()

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

	err := runCommand(cmd, cmdArgs)
	if err != nil {
		logger.Error(err)
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
	case "pairing-key":
		return runPairingKey(argv)
	case "ca":
		return runCA(argv)
	case "node":
		return runNode(argv)
	case "org":
		return runOrg(argv)
	}

	return fmt.Errorf("%s is not a pki.io command. See 'pki.io help'", cmd)
}

// Initialize logging from command arguments.
func initLogging(args map[string]interface{}) {
	if loggingConfig, ok := args["--logging"].(string); ok {
		var err error
		logger, err = seelog.LoggerFromConfigAsFile(loggingConfig)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize logging: %s", err))
		}
	}
}

func notImpl() (err error) {
	return errors.New("Not Implemented ...yet")
}
