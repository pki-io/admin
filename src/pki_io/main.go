package main

import (
    "fmt"
    "pki_io/document"
)

func main() {
    if doc, err := document.NewCA(nil); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("Source is: ", doc.Data.Options.Source)
        doc.Data.Options.Source = "aabbccddeeff00112233445566778899"
        fmt.Println("Source is now: ", doc.Data.Options.Source)
    }
}
