package main

import (
    //"fmt"
    "log"
    "pki_io/document"
)

func main() {
    doc := document.Document{}
    if err := doc.Load(`{"version":1,"type":"test-message","options":[],"body":{}}`); err != nil {
      log.Fatal(err)
    }
    if err := doc.Validate(); err != nil {
      log.Fatal(err)
    }
}
