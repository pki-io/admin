# OSX makefile
default: all
all:
	go build pki.io.go runAPI.go runClient.go runCA.go runEntity.go runAdmin.go
clean:
	rm pki.io
