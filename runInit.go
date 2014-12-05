package main

import (
    "fmt"
    "os"
    "pki.io/fs"
    "pki.io/entity"
    "pki.io/crypto"
    "pki.io/document"
    "encoding/hex"
)

func runInit(argv map[string]interface{}) (err error) {

    /**************************************************************************************************
    * Initialise the environment
    **************************************************************************************************/
    fmt.Println("Creating Org entity")
    org, err := entity.New(nil)
    if err != nil {
      panic(fmt.Sprintf("Could not create org entity: %s", err.Error()))
    }

    org.Data.Body.Id = hex.EncodeToString(crypto.RandomBytes(16))
    org.Data.Body.Name = argv["--org"].(string)

    currentDir, err := os.Getwd()
    if err != nil {
      panic(fmt.Sprintf("Could not get current directory: %s", err.Error()))
    }
    fsAPI, _ := fs.NewAPI(org.Data.Body.Name, currentDir)


    /**************************************************************************************************
    * Create the admin
    **************************************************************************************************/

    fmt.Println("Creating Admin entity")
    admin, err := entity.New(nil)
    if err != nil {
      panic(fmt.Sprintf("Could not create admin entity: %s", err.Error()))
    }

    admin.Data.Body.Id = hex.EncodeToString(crypto.RandomBytes(16))
    admin.Data.Body.Name = argv["--admin"].(string)
    err = admin.GenerateKeys()
    if err != nil {
      panic(fmt.Sprintf("Could not generate admin keys: %s", err.Error()))
    }

    fsAPI.Id = admin.Data.Body.Id

    fmt.Println("Saving admin")
    adminJson, err := admin.Dump()
    if err != nil {
        panic(fmt.Sprintf("Could not dump admin: %s", err.Error()))
    }

    if err := fsAPI.StorePrivate("admin", adminJson); err != nil {
        panic(fmt.Sprintf("Could not save admin data: %s", err.Error()))
    }

    /**************************************************************************************************
    * Org keys
    **************************************************************************************************/

    fmt.Println("Generating Org keys")
    err = org.GenerateKeys()
    if err != nil {
      panic(fmt.Sprintf("Could not generate org keys: %s", err.Error()))
    }

    fmt.Println("Creating public copy of org to save locally")
    publicOrg, err := org.Public()
    if err != nil {
      panic(fmt.Sprintf("Could get public org: %s", err.Error()))
    }
    publicOrgJson, err := publicOrg.Dump()
    if err != nil {
      panic(fmt.Sprintf("Could dump public org to json: %s", err.Error()))
    }

    fmt.Println("Creating public org container document")
    publicOrgContainer, err := document.NewContainer(nil)
    if err != nil {
        panic(fmt.Sprintf("Could not create container: %s", err.Error()))
    }
    publicOrgContainer.Data.Options.Source = publicOrg.Data.Body.Id
    publicOrgContainer.Data.Body = publicOrgJson

    fmt.Println("Signing public org container as admin")
    if err := admin.Sign(publicOrgContainer); err != nil  {
      panic(fmt.Sprintf("Could not sign container: %s", err.Error()))
    }

    fmt.Println("Dumping public org container")
    publicOrgContainerJson, err := publicOrgContainer.Dump()
    if err != nil {
        panic(fmt.Sprintf("Could not dump public org container json: %s", err.Error()))
    }

    fmt.Println("Saving org data")
    if err := fsAPI.StorePublic("org", publicOrgContainerJson); err != nil {
        panic(fmt.Sprintf("Could not save file: %s", err.Error()))
    }

    return nil
}
