package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/pki-io/pki.io/config"
	"github.com/pki-io/pki.io/document"
	"github.com/pki-io/pki.io/entity"
	"github.com/pki-io/pki.io/fs"
	"github.com/pki-io/pki.io/index"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	logger.Infof("Loading local configuration file: %s", configFile)
	conf := config.New(configFile)
	if err := conf.Load(); err != nil {
		panic(logger.Errorf("Could not load config file %s: %s", configFile, err))
	}
	return conf
}

func LoadAPI() *fs.FsAPI {
	logger.Info("Setting up the filesystem API")
	fsAPI, _ := fs.NewAPI(CurrentDir(), "") // we're in the name'd path
	return fsAPI
}

func LoadAdmin(fsAPI *fs.FsAPI, conf *config.Config) *entity.Entity {
	logger.Info("Reading local admin")
	fsAPI.Id = conf.Data.Admins[0].Id // need to override if required
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

func LoadOrgPrivate(fsAPI *fs.FsAPI, admin *entity.Entity, conf *config.Config) *entity.Entity {
	logger.Info("Loading private Org")
	fsAPI.Id = conf.Data.Org.Id

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

func LoadOrgPublic(fsAPI *fs.FsAPI, admin *entity.Entity, conf *config.Config) *entity.Entity {
	logger.Info("Loading public Org")
	fsAPI.Id = conf.Data.Org.Id

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

func LoadOrgIndex(fsAPI *fs.FsAPI, org *entity.Entity) *index.OrgIndex {
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

	indx, err := index.NewOrg(decryptedIndexJson)
	if err != nil {
		panic(logger.Errorf("Could not create indx: %s", err))
	}
	return indx
}

func SaveOrgIndex(fsAPI *fs.FsAPI, org *entity.Entity, indx *index.OrgIndex) {
	encryptedIndexContainer, err := org.EncryptThenSignString(indx.Dump(), nil)
	if err != nil {
		panic(logger.Errorf("Could not encrypt and sign index: %s", err))
	}
	if err := fsAPI.SendPrivate(org.Data.Body.Id, "index", encryptedIndexContainer.Dump()); err != nil {
		panic(logger.Errorf("Could not save encrypted: %s", err))
	}
}

type ExportFile struct {
	Name    string
	Mode    int64
	Owner   int64
	Group   int64
	Content []byte
}

func TarGZ(files []ExportFile) ([]byte, error) {
	tarBuffer := new(bytes.Buffer)
	tarWriter := tar.NewWriter(tarBuffer)

	for _, file := range files {
		header := &tar.Header{
			Name:    file.Name,
			Mode:    int64(file.Mode),
			Size:    int64(len(file.Content)),
			ModTime: time.Now(),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, err
		}
		if _, err := tarWriter.Write(file.Content); err != nil {
			return nil, err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, err
	}

	zipBuffer := new(bytes.Buffer)
	zipWriter := gzip.NewWriter(zipBuffer)
	zipWriter.Write(tarBuffer.Bytes())
	zipWriter.Close()

	return zipBuffer.Bytes(), nil
}

func Export(files []ExportFile, outFile string) {
	tarGz, err := TarGZ(files)
	if err != nil {
		panic(logger.Errorf("Couldn't tar.gz the files: %s", err))
	}

	if outFile == "-" {
		os.Stdout.Write(tarGz)
	} else {
		// Write  to file
		if err := ioutil.WriteFile(outFile, tarGz, 0600); err != nil {
			panic(logger.Errorf("Couldn't write export file: %s", err))
		}
	}

}
