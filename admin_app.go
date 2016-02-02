// ThreatSpec package main
package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/olekukonko/tablewriter"
	"github.com/pki-io/controller"
	"os"
)

type AdminApp struct {
	env *controller.Environment
}

func NewAdminApp() *AdminApp {
	app := new(AdminApp)
	app.env = controller.NewEnvironment()
	controller.UseLogger(logger)
	return app
}

func (app *AdminApp) Exit() {
	logger.Flush()
	cli.Exit(0)
}

func (app *AdminApp) Fatal(err error) {
	congrats := `*************************************************
*                CONGRATULATIONS                *
*************************************************

You may have just found a bug in pki.io :)

Please try reproducing the problem using the trace log level by running:

    pki.io -l trace COMMAND [args...]

Let us know of the issue by raising an issue on GitHub here: 

    https://github.com/pki-io/core/issues

Or by dropping an email to: dev@pki.io

If possible, please include this full log output, including the error
and anything else relevant like what command you ran.

Many thanks,
The pki.io team`

	logger.Critical(err)
	fmt.Println(congrats)
	cli.Exit(1)
}

func (app *AdminApp) NewTable() *tablewriter.Table {
	logger.Debug("creating table")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	return table
}

func (app *AdminApp) RenderTable(table *tablewriter.Table) {
	logger.Debug("rendering output")
	logger.Flush()
	table.Render()
}
