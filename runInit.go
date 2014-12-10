package main

import (
    "fmt"
    "os"
    "pki.io/fs"
    "pki.io/entity"
    "pki.io/crypto"
    "pki.io/config"
    "encoding/hex"
    "path/filepath"
)

func runInit(argv map[string]interface{}) (err error) {
    /**************************************************************************************************
    * Load the config file
    **************************************************************************************************/

    globalConf := config.Global(".pki.io.conf")
    if err := globalConf.Load(); err != nil {
        panic("Could not load config file")
    }

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

    // Save the config file
    globalConf.AddOrg(org.Data.Body.Name, fsAPI.Path)
    if err := globalConf.Save(); err != nil {
        panic(fmt.Sprintf("Could not save config file: %s", err.Error()))
    }

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
    adminJson := admin.Dump()
    if err != nil {
        panic(fmt.Sprintf("Could not dump admin: %s", err.Error()))
    }

    if err := fsAPI.StorePrivate("admin", adminJson); err != nil {
        panic(fmt.Sprintf("Could not save admin data: %s", err.Error()))
    }

    // Save the org config
    orgFile := filepath.Join(fsAPI.Path, "org.conf")
    orgConfig := config.Org(orgFile)
    if err := orgConfig.Load(); err != nil {
        panic(fmt.Sprintf("Could not load org config: %s", err.Error()))
    }
    orgConfig.Data.OrgId = org.Data.Body.Id
    orgConfig.Data.AdminId = admin.Data.Body.Id
    if err := orgConfig.Save(); err != nil {
        panic(fmt.Sprintf("Could not save org config: %s", err.Error()))
    }

    /**************************************************************************************************
    * Org keys
    **************************************************************************************************/

    fmt.Println("Generating Org keys")
    err = org.GenerateKeys()
    if err != nil {
      panic(fmt.Sprintf("Could not generate org keys: %s", err.Error()))
    }

    // Public keys
    fmt.Println("Creating public copy of org to save locally")
    publicOrg, err := org.Public()
    if err != nil {
      panic(fmt.Sprintf("Could get public org: %s", err.Error()))
    }
    publicOrgContainer, err := admin.SignString(publicOrg.Dump())
    if err != nil {
        panic(fmt.Sprintf("Could not sign public org: %s", err.Error()))
    }

    fmt.Println("Saving org data")
    if err := fsAPI.StorePublic("org", publicOrgContainer.Dump()); err != nil {
        panic(fmt.Sprintf("Could not save file: %s", err.Error()))
    }

    // Private keys
    container, err := admin.EncryptThenSignString(org.Dump(), nil)
    if err != nil {
      panic(fmt.Sprintf("Could encrypt org: %s", err.Error()))
    }

    if err := fsAPI.StorePrivate("org", container.Dump()); err != nil {
      panic(fmt.Sprintf("Could not store container to json: %s", err.Error()))
    }
    return nil
}
