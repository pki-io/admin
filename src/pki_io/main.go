package main

import (
    "fmt"
    "pki_io/document"
)

func main() {
    if doc, err := document.New(`{}`); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(doc)
    }

    if doc, err := document.NewCA(`{"version":1,"type":"ca-document","options":{"source":"123","signature-mode":""},"body":{}}`); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(doc)
    }

    if doc, err := document.NewCA(nil); err != nil {
        fmt.Println(err)
    } else {
        doc.Options()["source"] = "333"
        fmt.Println(doc)
    }
}
