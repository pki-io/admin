package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	"os"
)

var logger log.LoggerInterface

type CustomReceiver struct{}

// The next const isn't required anymore but leaving here for now as we might need it when we try
// to get config working again.
const defaultLoggingConfig string = `
<seelog minlevel="%s" maxlevel="error">
    <outputs formatid="raw">
        <custom name="stderr"/>
    </outputs>
    <formats>
        <format id="raw" format="%%Msg%%n"/>
    </formats>
</seelog>
`

func checkLogFatal(format string, a ...interface{}) {
	if len(a) > 0 && a[len(a)-1] == nil {
		return
	}
	fmt.Fprintf(os.Stderr, format, a...)
	panic("...")
	os.Exit(1)
}

// Initialize logging from command arguments.
func initLogging(level, configFile string) {
	var err error
	log.RegisterReceiver("stderr", &CustomReceiver{})

	if configFile == "" {
		// TODO - better validation?
		if level != "error" && level != "warn" && level != "info" && level != "debug" && level != "trace" {
			checkLogFatal("Invalid log level: %s", level)
		}

		logger, err = log.LoggerFromConfigAsString(fmt.Sprintf(defaultLoggingConfig, level))
		checkLogFatal("Failed to load default logging configuration: %s", err)
	} else {
		logger, err = log.LoggerFromConfigAsFile(configFile)
		checkLogFatal("Failed to initialize custom logging file %s: %s", configFile, err)
	}
}

func (ar *CustomReceiver) ReceiveMessage(message string, level log.LogLevel, context log.LogContextInterface) error {
	fmt.Fprintf(os.Stderr, "%s", message)
	return nil
}

func (ar *CustomReceiver) AfterParse(initArgs log.CustomReceiverInitArgs) error {
	return nil
}

func (ar *CustomReceiver) Flush() {

}

func (ar *CustomReceiver) Close() error {
	return nil
}
