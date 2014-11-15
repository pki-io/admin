package main

import (
    "fmt"
    // http://docopt.org/
    "github.com/docopt/docopt-go"
)
func main() {
    usage := `
Usage:
    TBC
Options:
    -h --help   Show this screen
    `

   arguments, _ := docopt.Parse(usage, nil, true, "pki.io admin-CLI", false)
   fmt.Println(arguments)
}
