package document

import (
    "fmt"
    "encoding/json"
)
type Documenter interface {
    Version()
    Type()
    Options()
    Body()
}

type document struct {
    Version int
    Type string
    Options interface{}
    Body interface{}
}

type Document struct {
    document
}

func New(json_data string) (*Document, error) {
    d := document{}
    if err := json.Unmarshal([]byte(json_data), &d); err != nil {
        return nil, err
    }
    //doc := Document{d}
    return &Document{d}, nil
}

func (doc *Document) Version() int {
    return doc.document.Version
}

func (doc *Document) Type() string {
    return doc.document.Type
}

func (doc *Document) Options() interface{} {
    return doc.document.Options
}

func (doc *Document) Body() interface{} {
    return doc.document.Body
}

type CADocument struct {
    document
}

func NewCA(data interface{}) (*CADocument, error) {
    d := document{}

    var json_data []byte
    switch t := data.(type) {
        case []byte:
            json_data = data.([]byte)
        case string:
            json_data = []byte(data.(string))
        case nil:
            json_data = []byte(`{"version":1,"type":"ca-document","options":{"source":"","signature-mode":""},"body":{}}`)
        default:
            return nil, fmt.Errorf("Invalid input type: %T", t)
    }

    if err := json.Unmarshal(json_data, &d); err != nil {
        return nil, err
    }

    if t := d.Type; t != "ca-document" {
      return nil, fmt.Errorf("Invalid type for CADocument: %s", t)
    }

    switch t := d.Options.(type) {
      case map[string]interface{}:
      default:
      return nil, fmt.Errorf("Invalid options type. Must be a hash: %T", t)
    }

    if _, ok := d.Options.(map[string]interface{})["source"]; ! ok {
      return nil, fmt.Errorf("Options source missing")
    }

    if _, ok := d.Options.(map[string]interface{})["signature-mode"]; ! ok {
      return nil, fmt.Errorf("Options signature-mode missing")
    }

    switch t := d.Body.(type) {
      case map[string]interface{}:
      default:
      return nil, fmt.Errorf("Invalid body type. Must be a hash: %T", t)
    }

    return &CADocument{d}, nil
}

func (doc *CADocument) Version() int {
    return doc.document.Version
}

func (doc *CADocument) Type() string {
    return doc.document.Type
}

func (doc *CADocument) Options() map[string]interface{} {
    return doc.document.Options.(map[string]interface{})
}

func (doc *CADocument) Body() map[string]interface{} {
    return doc.document.Body.(map[string]interface{})
}
