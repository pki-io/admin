package main

import (
    "fmt"
    "pki_io/document"
)

func main() {
    doc, err := document.NewCA(nil)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("Source is: ", doc.Data.Options.Source)
        doc.Data.Options.Source = "aabbccddeeff00112233445566778899"
        newDoc, _ := doc.Json()
        fmt.Println("Document is now: ", newDoc)
    }
}
