package document

import (
    "testing"
)

func TestDocumentEmpty(t *testing.T) {
    doc := Document{}
    if err := doc.Load(`{}`); err != nil {
      t.Errorf("Failed to load document: %s", err)
    }

    if err := doc.Validate(); err == nil {
      t.Errorf("Should not be able to validate empty document")
    }
}

func TestDocumentValid(t *testing.T) {
    doc := Document{}
    if err := doc.Load(`{"version":1,"type":"test","options":"test","body":"test"}`); err != nil {
      t.Errorf("Failed to load document: %s", err)
    }

    if err := doc.Validate(); err != nil {
      t.Errorf("Failed to validate document: %s", err)
    }

    if doc.Version != 1 {
      t.Errorf("Document version not as set")
    }

    if doc.Type != "test" {
      t.Errorf("Document type not as set")
    }

    if doc.Options != "test" {
      t.Errorf("Document options not as set")
    }

    if doc.Body != "test" {
      t.Errorf("Document body not as set")
    }
}

func TestDocumentInValidVersion(t *testing.T) {
    doc := Document{}
    if err := doc.Load(`{"version":"a string","type":"test","options":"test","body":"test"}`); err == nil {
      t.Errorf("Should not be able to validate invalid version")
    } else if err.Error() != "json: cannot unmarshal string into Go value of type int" {
      t.Errorf("Unexpected error loading invalid document", err)
    }

    if err := doc.Validate(); err == nil {
      t.Errorf("Should not be able to validate invalid version")
    } else if err.Error() != "Version not set" {
      t.Errorf("Unexpected error validating document", err)
    }
}
