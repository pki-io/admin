package main

import (
    "fmt"
    "os"
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

    currentDir, err := os.Getwd()
    if err != nil {
        panic(fmt.Sprintf("Could not get current directory: %s", err.Error()))
    }

    configFile := filepath.Join(currentDir, "pki.io.conf")
    conf := config.New(configFile)
    if err := conf.Load(); err != nil {
        panic(fmt.Sprintf("Could not load config file %s: %s", configFile, err.Error()))
    }

    fsAPI, _ := fs.NewAPI(currentDir, "") // we're in the name'd path
    fsAPI.Id = conf.Data.AdminId

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
