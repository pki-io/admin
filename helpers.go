package main

import (
	"fmt"
	"github.com/mitchellh/packer/common/uuid"
	"os"
	"path/filepath"
	"pki.io/config"
	"pki.io/document"
	"pki.io/entity"
	"pki.io/fs"
)

func NewID() string {
	return uuid.TimeOrderedUUID()
}

func CurrentDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Could not get current directory: %s", err.Error()))
	}
	return currentDir
}

func LoadConfig() *config.Config {

	configFile := filepath.Join(CurrentDir(), "pki.io.conf")
	conf := config.New(configFile)
	if err := conf.Load(); err != nil {
		panic(fmt.Sprintf("Could not load config file %s: %s", configFile, err.Error()))
	}
	return conf
}

func LoadAPI(conf *config.Config) *fs.FsAPI {
	fsAPI, _ := fs.NewAPI(CurrentDir(), "") // we're in the name'd path
	fsAPI.Id = conf.Data.Admins[0].Id       // need to override if required
	return fsAPI
}

func LoadAdmin(fsAPI *fs.FsAPI) *entity.Entity {
	adminJson, err := fsAPI.ReadLocal("admin")
	if err != nil {
		panic(fmt.Sprintf("Could not load admin data: %s", err.Error()))
	}
	admin, err := entity.New(adminJson)
	if err != nil {
		panic(fmt.Sprintf("Could not create admin entity: %s", err.Error()))
	}
	return admin
}

func LoadOrgPrivate(fsAPI *fs.FsAPI, admin *entity.Entity) *entity.Entity {
	orgJson, err := fsAPI.LoadPrivate("org")
	if err != nil {
		panic(fmt.Sprintf("Could not load org data: %s", err.Error()))
	}
	orgContainer, err := document.NewContainer(orgJson)
	if err != nil {
		panic(fmt.Sprintf("Could not load org container: %s", err.Error()))
	}

	if err := admin.Verify(orgContainer); err != nil {
		panic(fmt.Sprintf("Could not verify org: %s", err.Error()))
	}

	decryptedOrgJson, err := admin.Decrypt(orgContainer)
	if err != nil {
		panic(fmt.Sprintf("Could not decrypt container: %s", err.Error()))
	}

	org, err := entity.New(decryptedOrgJson)
	if err != nil {
		panic(fmt.Sprintf("Could not create org entity: %s", err.Error()))
	}
	return org
}

func LoadOrgPublic(fsAPI *fs.FsAPI, admin *entity.Entity) *entity.Entity {
	orgJson, err := fsAPI.LoadPublic("org")
	if err != nil {
		panic(fmt.Sprintf("Could not load org data: %s", err.Error()))
	}
	orgContainer, err := document.NewContainer(orgJson)
	if err != nil {
		panic(fmt.Sprintf("Could not load org container: %s", err.Error()))
	}

	if err := admin.Verify(orgContainer); err != nil {
		panic(fmt.Sprintf("Could not verify org: %s", err.Error()))
	}

	org, err := entity.New(orgContainer.Data.Body)
	if err != nil {
		panic(fmt.Sprintf("Could not create org entity: %s", err.Error()))
	}
	return org
}
