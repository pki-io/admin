package document

import (
    "encoding/json"
    "io"
    "strings"
    "fmt"
    //"errors"
)

type Document struct {
    Version int
    Type string
    Options interface{}
    Body interface{}
}

func (doc *Document) Load(data string) (err error) {
    dec := json.NewDecoder(strings.NewReader(data))
    if err := dec.Decode(doc); err == io.EOF {
        return nil
    } else if err != nil {
        return err
    }
    return err
}

func (doc *Document) Validate() (err error) {
    switch v := doc.Options.(type) {
        case string:
        case []interface{}:
        case map[string]interface{}:
        default:
          return fmt.Errorf("Invalid type for Options: %T", v)
    }

    switch v := doc.Body.(type) {
        case string:
        case []interface{}:
        case map[string]interface{}:
        default:
          return fmt.Errorf("Invalid type for Body: %T", v)
    }
    return
}

type CADocument struct {
    Document // Anonymous field Document
}

func (doc *CADocument) Validate() (err error) {
    switch v:= doc.Options.(type) {
        case map[string]interface{}:
        default:
          return fmt.Errorf("Invalid type for CADocument Options: %T", v)
    }

    switch v:= doc.Body.(type) {
        case map[string]interface{}:
        default:
          return fmt.Errorf("Invalid type for CADocument Body: %T", v)
    }
    return
}
