package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

type AdminApp struct {
	env *Environment
}

func NewAdminApp() *AdminApp {
	app := new(AdminApp)

	app.env = new(Environment)

	return app
}

func (app *AdminApp) Exit() {
	logger.Flush()
	os.Exit(0)
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
The pki.io team

Your error was: %s`
	logger.Critical(fmt.Sprintf(congrats, err))
	os.Exit(1)
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
