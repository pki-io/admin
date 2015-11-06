// ThreatSpec package main
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
	"time"
)

func NewID() string {
	idBytes, _ := crypto.RandomBytes(16)
	return hex.EncodeToString(idBytes)
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
