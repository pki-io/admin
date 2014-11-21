package document

import (
    "fmt"
    "strings"
    "encoding/json"
)

type document struct {
    Version int
    Type string
    Options interface{}
    Body interface{}
}

func (doc *document) parseJson(data interface{}) (error) {
    var json_data []byte
    switch t := data.(type) {
        case []byte:
            json_data = data.([]byte)
        case string:
            json_data = []byte(data.(string))
        case nil:
            json_data = []byte(`{"version":1,"type":"ca-document","options":{"source":"","signature-mode":""},"body":{}}`)
        default:
            return fmt.Errorf("Invalid input type: %T", t)
    }
    return json.Unmarshal(json_data, doc)
}

type Document struct {
    document
}

func New(data string) (*Document, error) {
    d := document{}
    if err := d.parseJson(data); err != nil {
        return nil, err
    }
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
    if err := d.parseJson(data); err != nil {
        return nil, err
    }

    doc := CADocument{d}
    if err := doc.Validate(); err != nil {
        return nil, err
    } else {
        return &doc, nil
    }
}

func (doc *CADocument) Validate() (err error) {
    var errors []string

    if t := doc.document.Type; t != "ca-document" {
      errors = append(errors, fmt.Sprintf("Invalid type for CADocument: %s", t))
    }

    switch t := doc.document.Options.(type) {
        case map[string]interface{}:
        default:
            errors = append(errors, fmt.Sprintf("Invalid options type. Must be a map: %T", t))
    }

    if _, ok := doc.document.Options.(map[string]interface{})["source"]; ! ok {
        errors = append(errors, fmt.Sprintf("Options source missing"))
    }

    if _, ok := doc.document.Options.(map[string]interface{})["signature-mode"]; ! ok {
        errors = append(errors, fmt.Sprintf("Options signature-mode missing"))
    }

    switch t := doc.document.Body.(type) {
        case map[string]interface{}:
        default:
            errors = append(errors, fmt.Sprintf("Invalid body type. Must be a hash: %T", t))
    }

    if len(errors) > 0 {
        err = fmt.Errorf("Could not validate: %s", strings.Join(errors, ", "))
    }
    return
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

func (doc *CADocument) Json() (string, error) {
    if err := doc.Validate(); err != nil {
        return "", err
    } else if b, err := json.Marshal(doc.document); err != nil {
        return "", err
    } else {
        return string(b[:]), nil
    }
}

func (doc *CADocument) String() string {
    str, _ := doc.Json()
    return str
}
