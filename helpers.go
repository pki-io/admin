package main

import (
	"github.com/mitchellh/packer/common/uuid"
	"os"
	"path/filepath"
	"github.com/pki-io/pki.io/config"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/entity"
	"github.com/pki-io/pki.io/fs"
	"github.com/pki-io/pki.io/index"
	"strings"
)

func NewID() string {
	return uuid.TimeOrderedUUID()
}

func CurrentDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(logger.Errorf("Could not get current directory: %s", err))
	}
	return currentDir
}

func LoadConfig() *config.Config {

	configFile := filepath.Join(CurrentDir(), "pki.io.conf")
	conf := config.New(configFile)
	if err := conf.Load(); err != nil {
		panic(logger.Errorf("Could not load config file %s: %s", configFile, err))
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
		panic(logger.Errorf("Could not load admin data: %s", err))
	}
	admin, err := entity.New(adminJson)
	if err != nil {
		panic(logger.Errorf("Could not create admin entity: %s", err))
	}
	return admin
}

func LoadOrgPrivate(fsAPI *fs.FsAPI, admin *entity.Entity) *entity.Entity {
	orgJson, err := fsAPI.LoadPrivate("org")
	if err != nil {
		panic(logger.Errorf("Could not load org data: %s", err))
	}
	orgContainer, err := document.NewContainer(orgJson)
	if err != nil {
		panic(logger.Errorf("Could not load org container: %s", err))
	}

	if err := admin.Verify(orgContainer); err != nil {
		panic(logger.Errorf("Could not verify org: %s", err))
	}

	decryptedOrgJson, err := admin.Decrypt(orgContainer)
	if err != nil {
		panic(logger.Errorf("Could not decrypt container: %s", err))
	}

	org, err := entity.New(decryptedOrgJson)
	if err != nil {
		panic(logger.Errorf("Could not create org entity: %s", err))
	}
	return org
}

func LoadOrgPublic(fsAPI *fs.FsAPI, admin *entity.Entity) *entity.Entity {
	orgJson, err := fsAPI.LoadPublic("org")
	if err != nil {
		panic(logger.Errorf("Could not load org data: %s", err))
	}
	orgContainer, err := document.NewContainer(orgJson)
	if err != nil {
		panic(logger.Errorf("Could not load org container: %s", err))
	}

	if err := admin.Verify(orgContainer); err != nil {
		panic(logger.Errorf("Could not verify org: %s", err))
	}

	org, err := entity.New(orgContainer.Data.Body)
	if err != nil {
		panic(logger.Errorf("Could not create org entity: %s", err))
	}
	return org
}

func ParseTags(tagString string) []string {
	tags := strings.Split(tagString, ",")
	for i, e := range tags {
		tags[i] = strings.TrimSpace(strings.ToLower(e))
	}
	return tags
}

func LoadIndex(fsAPI *fs.FsAPI, org *entity.Entity) *index.Index {
	indexJson, err := fsAPI.GetPrivate(org.Data.Body.Id, "index")
	if err != nil {
		panic(logger.Errorf("Could not load index data: %s", err))

	}
	indexContainer, err := document.NewContainer(indexJson)
	if err != nil {
		panic(logger.Errorf("Could not load index container: %s", err))
	}

	if err := org.Verify(indexContainer); err != nil {
		panic(logger.Errorf("Could not verify index: %s", err))
	}

	decryptedIndexJson, err := org.Decrypt(indexContainer)
	if err != nil {
		panic(logger.Errorf("Could not decrypt container: %s", err))
	}

	indx, err := index.New(decryptedIndexJson)
	if err != nil {
		panic(logger.Errorf("Could not create indx: %s", err))
	}
	return indx
}

func SaveIndex(fsAPI *fs.FsAPI, org *entity.Entity, indx *index.Index) {
	encryptedIndexContainer, err := org.EncryptThenSignString(indx.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt and sign index: %s", err))
	}
	if err := fsAPI.SendPrivate(org.Data.Body.Id, "index", encryptedIndexContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save encrypted: %s", err))
	}
}
