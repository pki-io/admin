package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"github.com/pki-io/core/crypto"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func NewID() string {
	idBytes, err := crypto.RandomBytes(16)
	fmt.Println(err)
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
		argInt, err := strconv.ParseInt(arg.(string), 10, 64)
		fmt.Println(err)
		return int(argInt)
	case nil:
		return def.(int)
	case bool:
		return def.(int)
	default:
		// Never gets to the next line
		fmt.Println(t)
		return 0
	}
}

func ArgString(arg interface{}, def interface{}) string {
	switch t := arg.(type) {
	case string:
		return arg.(string)
	case nil:
		return def.(string)
	case bool:
		return def.(string)
	default:
		// Never gets to the next line
		fmt.Println(t)
		return ""
	}
}

func ArgBool(arg interface{}, def interface{}) bool {
	switch t := arg.(type) {
	case string:
		return arg.(string) == "true"
	case bool:
		return arg.(bool)
	default:
		fmt.Println(t)
		return false
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
	fmt.Println(err)

	if outFile == "-" {
		os.Stdout.Write(tarGz)
	} else {
		// Write  to file
		err := ioutil.WriteFile(outFile, tarGz, 0600)
		fmt.Println(err)
	}

}
