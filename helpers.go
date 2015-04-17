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

func checkUserFatal(format string, a ...interface{}) {
	if len(a) > 0 && a[len(a)-1] == nil {
		return
	}
	logger.Errorf(format, a...)
	os.Exit(1)
}

func checkAppFatal(format string, a ...interface{}) {
	if len(a) > 0 && a[len(a)-1] == nil {
		return
	}

	bigError := "*************************************************\n" +
		"*                CONGRATULATIONS                *\n" +
		"*************************************************\n\n" +
		"You may have just found a bug in pki.io :)\n\n" +
		"Please let us know by raising an issue on GitHub here: https://github.com/pki-io/core/issues\n\n" +
		"Or by dropping an email to: dev@pki.io\n\n" +
		"If possible, please include this full error message, including the below panic,\n" +
		"and anything else relevant like what command you ran.\n\n" +
		"Many thanks,\n" +
		"The pki.io team\n\n" +
		"The error was: " + format + "\n\n"

	fmt.Println(logger)
	logger.Errorf(bigError, a...)
	panic("...")
}

func NewID() string {
	idBytes, err := crypto.RandomBytes(16)
	checkAppFatal("Couldn't get random bytes: %s", err)
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
		checkAppFatal("Couldn't convert to int: %s", err)
		return int(argInt)
	case nil:
		return def.(int)
	case bool:
		return def.(int)
	default:
		checkAppFatal("Wrong arg type: %T", t)
		// Never gets to the next line
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
		checkAppFatal("Wrong arg type %T for arg %s", t, arg)
		// Never gets to the next line
		return ""
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
	checkAppFatal("Couldn't tar.gz the files: %s", err)

	if outFile == "-" {
		os.Stdout.Write(tarGz)
	} else {
		// Write  to file
		err := ioutil.WriteFile(outFile, tarGz, 0600)
		checkAppFatal("Couldn't write export file: %s", err)
	}

}
