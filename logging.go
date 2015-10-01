package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	"os"
)

type CustomReceiver struct{}

var logger log.LoggerInterface

// The next const isn't required anymore but leaving here for now as we might need it when we try
// to get config working again.
//<formats>
//    <format id="raw" format="%%Msg%%n"/>
//</formats>
const defaultLoggingConfig string = `
<seelog minlevel="info" maxlevel="critical">
  <outputs>
    <filter levels="info">
      <custom name="stderr" formatid="default"/>
    </filter>
    <filter levels="warn,error,critical">
      <custom name="stderr" formatid="error"/>
    </filter>
  </outputs>
  <formats>
    <format id="default" format="%Msg%n"/>
    <format id="error" format="%LEVEL %Msg%n"/>
  </formats>
</seelog>
`

const verboseLoggingConfig string = `
<seelog minlevel="%s" maxlevel="critical">
  <outputs formatid="verbose">
    <custom name="stderr"/>
  </outputs>
  <formats>
    <format id="verbose" format="%%Time %%LEVEL [%%FuncShort @ %%File.%%Line] %%Msg%%n"/>
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

func initLogging(level, configFile string) log.LoggerInterface {
	var err error
	log.RegisterReceiver("stderr", &CustomReceiver{})

	if configFile == "" {
		// TODO - better validation?
		if level != "error" && level != "warn" && level != "info" && level != "debug" && level != "trace" {
			checkLogFatal("Invalid log level: %s", level)
		}

		if level == "info" {
			logger, err = log.LoggerFromConfigAsString(defaultLoggingConfig)
			checkLogFatal("Failed to load default logging configuration: %s", err)
		} else {
			logger, err = log.LoggerFromConfigAsString(fmt.Sprintf(verboseLoggingConfig, level))
			checkLogFatal("Failed to load default logging configuration: %s", err)
		}
	} else {
		logger, err = log.LoggerFromConfigAsFile(configFile)
		checkLogFatal("Failed to initialize custom logging file %s: %s", configFile, err)
	}

	//defer logger.Close()

	return logger
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
