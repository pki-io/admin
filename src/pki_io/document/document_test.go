package document

import (
    "testing"
)

func TestDocumentEmpty(t *testing.T) {
    doc := Document{}
    if err := doc.Load(`{}`); err != nil {
      t.Errorf("Failed to load empty document: %s", err)
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

func TestDocumentInvalidVersion(t *testing.T) {
    doc := Document{}
    if err := doc.Load(`{"version":"a string","type":"test","options":"test","body":"test"}`); err == nil {
      t.Errorf("Should not be able to load invalid version")
    } else if err_string := err.Error(); err_string != "json: cannot unmarshal string into Go value of type int" {
      t.Errorf("Unexpected error loading invalid document", err_string)
    }
}

func TestDocumentInvalidType(t *testing.T) {
    doc := Document{}
    if err := doc.Load(`{"version":1,"type":1,"options":"test","body":"test"}`); err == nil {
      t.Errorf("Should not be able to load invalid type")
    } else if err_string := err.Error(); err_string != "json: cannot unmarshal number into Go value of type string" {
      t.Errorf("Unexpected error loading invalid document", err_string)
    }
}

func TestDocumentInvalidOptions(t *testing.T) {
    doc := Document{}
    if err := doc.Load(`{"version":1,"type":"test","options":1,"body":"test"}`); err != nil {
      t.Errorf("Should be able to load invalid options: %s", err)
    }

    if err := doc.Validate(); err == nil {
      t.Errorf("Should not be able to validate invalid options")
    } else if err_string := err.Error(); err_string != "Invalid type for Options: float64" {
      t.Errorf("Unexpected error validating options: %s", err_string)
    }
}

func TestDocumentInvalidBody(t *testing.T) {
    doc := Document{}
    if err := doc.Load(`{"version":1,"type":"test","options":"","body":1}`); err != nil {
      t.Errorf("Should be able to load invalid body: %s", err)
    }

    if err := doc.Validate(); err == nil {
      t.Errorf("Should not be able to validate invalid body")
    } else if err_string := err.Error(); err_string != "Invalid type for Body: float64" {
      t.Errorf("Unexpected error validating body: %s", err_string)
    }
}






func TestCADocumentEmpty(t *testing.T) {
    doc := CADocument{Document{}}
    if err := doc.Load(`{}`); err != nil {
        t.Errorf("Failed to load empty document: %s", err)
    }

    if err := doc.Validate(); err == nil {
        t.Errorf("Should not be able to validate invalid body")
    } else if err_string := err.Error(); err_string != "Invalid type for CADocument Options: <nil>" {
        t.Errorf("Unexpected error validating body: %s", err_string)
    }
}

func TestCADocumentValid(t *testing.T) {
    doc := CADocument{Document{}}
    test_document := `
      {
        "version": 1,
        "type": "ca-document",
        "options": {
          "source": "test",
          "signature_mode": "test",
          "signature": "test"
        },
        "body": {
          "tags": ["test"],
          "certificate": "test",
          "private-key": "test"
        }  
      }
      `
    if err := doc.Load(test_document); err != nil {
        t.Errorf("Failed to load document: %s", err)
    }

    if err := doc.Validate(); err != nil {
        t.Errorf("Failed to validate document: %s", err)
    }

    if doc.Version != 1 {
        t.Errorf("CADocument version not as set")
    }

    if doc.Type != "ca-document" {
        t.Errorf("CADocument type not as set")
    }

    // Uhmm, really?
    if val, ok := doc.Options.(map[string]interface{})["source"]; ! ok {
        t.Errorf("CADocument missing source option")
    } else if val != "test" {
        t.Errorf("CADocument options source not as set")
    }
    /*

    src/pki_io/document/document_test.go:121: invalid argument doc.Document.Options (type interface {}) for len

    if len(doc.Options) != 0 {
      t.Errorf("Document options not as set")
    }

    if len(doc.Body) != 0 {
      t.Errorf("Document body not as set")
    }
    */
}


