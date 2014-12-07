package main

import (
    "fmt"
    "pki.io/config"
    "pki.io/fs"
    "pki.io/entity"
    "pki.io/document"
    "path/filepath"
)
/*
import (
    "os"
    "pki.io/crypto"
    "encoding/hex"
)*/

func orgShow(argv map[string]interface{}) (err error) {
    globalConf := config.Global(".pki.io.conf")
    globalConf.Load()

    if len(globalConf.Data.Org) == 0 {
        return fmt.Errorf("No orgs defined")
    }
    // Do default search here and probably move to help function
    orgConfigFile := filepath.Join(globalConf.Data.Org[0].Path, "org.conf")
    orgConf := config.Org(orgConfigFile)
    orgConf.Load()

    fsAPI, _ := fs.NewAPI("", globalConf.Data.Org[0].Path)
    fsAPI.Id = orgConf.Data.AdminId

    adminJson, err := fsAPI.LoadPrivate("admin")
    if err != nil {
      panic(fmt.Sprintf("Could not load admin data: %s", err.Error()))
    }
    admin, err := entity.New(adminJson)
    if err != nil {
      panic(fmt.Sprintf("Could not create admin entity: %s", err.Error()))
    }

    orgJson, err := fsAPI.LoadPublic("org")
    if err != nil {
      panic(fmt.Sprintf("Could not load org data: %s", err.Error()))
    }
    orgContainer, err := document.NewContainer(orgJson)
    if err != nil {
      panic(fmt.Sprintf("Could not load org container: %s", err.Error()))
    }

    if err := admin.Verify(orgContainer); err != nil {
      panic(fmt.Sprintf("Could not load verify container: %s", err.Error()))
    }

    org, err := entity.New(orgContainer.Data.Body)
    if err != nil {
      panic(fmt.Sprintf("Could not create org entity: %s", err.Error()))
    }

    fmt.Printf("Name: %s\n", org.Data.Body.Name)
    fmt.Printf("Id: %s\n", org.Data.Body.Id)
    fmt.Printf("Public Signing Key:\n%s\n", org.Data.Body.PublicSigningKey)
    fmt.Printf("Public Encryption Key:\n%s\n", org.Data.Body.PublicEncryptionKey)

    return nil
}

func runOrg(argv map[string]interface{}) (err error) {
    if argv["show"].(bool) {
        orgShow(argv)
    }
    return nil
}
