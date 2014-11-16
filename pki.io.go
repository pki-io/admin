package main

import (
    "fmt"
    // http://docopt.org/
    "github.com/docopt/docopt-go"
)
func main() {
    usage := `pki.io
Usage:
    pki.io admin init ENTITY
    pki.io admin revoke ENTITY
    edmin-cli ntity new ID
    pki.io entity remove ID
    pki.io ca new NAME
    pki.io ca remove ID
    pki.io ca rotate ID
    pki.io ca freeze ID
    pki.io ca revoke ID
    pki.io client new IP
    pki.io client remove ID
    pki.io client ID
    pki.io client freeze ID
    pki.io client revoke ID
    pki.io api status
Options:
    -h --help   Show this screen
    -v --version Show version
    `

   arguments, _ := docopt.Parse(usage, nil, true, "pki.io", false)
   fmt.Println(arguments)
}
