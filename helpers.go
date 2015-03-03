package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"github.com/pki-io/core/crypto"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func NewID() string {
	idBytes, err := crypto.RandomBytes(16)
	if err != nil {
		panic(logger.Errorf("Couldn't get random bytes: %s", err))
	}
	return hex.EncodeToString(idBytes)
}

func ParseTags(tagString string) []string {
	tags := strings.Split(tagString, ",")
	for i, e := range tags {
		tags[i] = strings.TrimSpace(strings.ToLower(e))
	}
	return tags
}

func ArgInt(arg interface{}, def interface{}) int {
	switch t := arg.(type) {
	case string:
		if argInt, err := strconv.ParseInt(arg.(string), 10, 64); err != nil {
			panic(logger.Errorf("Couldn't convert to int: %s", err))
		} else {
			return int(argInt)
		}
	case nil:
		return def.(int)
	default:
		panic(logger.Errorf("Wrong arg type: %T", t))
	}
}

func ArgString(arg interface{}, def interface{}) string {
	switch t := arg.(type) {
	case string:
		return arg.(string)
	case nil:
		return def.(string)
	default:
		panic(logger.Errorf("Wrong arg type: %T", t))
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
